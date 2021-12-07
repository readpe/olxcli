// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package busfaults

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/readpe/goolx"
	"github.com/readpe/goolx/constants"
	"github.com/readpe/goolx/model"
	"github.com/spf13/cobra"
)

func init() {
	// Subcommand Flags.
	BFCmd.Flags().StringVarP(&fileFlag, "file", "f", "", "input *.olr file")
	BFCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "output file, if ending in '.csv' will output in csv format")
	BFCmd.Flags().StringVarP(&formatFlag, "format", "F", "tab", "output file format [csv]")
	BFCmd.Flags().StringVarP(&regexFlag, "expression", "e", "", "regular expression pattern")
	BFCmd.Flags().Float64Var(&vminFlag, "vmin", 0.0, "minimum voltage (kV)")
	BFCmd.Flags().Float64Var(&vmaxFlag, "vmax", 999.0, "maximum voltage (kV)")
	BFCmd.Flags().IntVarP(&areaFlag, "area", "a", 0, "area number")
	BFCmd.Flags().IntVarP(&zoneFlag, "zone", "z", 0, "zone number")
	BFCmd.Flags().Float64VarP(&r, "resistance", "r", 0.0, "fault resistance (Ohms)")
	BFCmd.Flags().Float64VarP(&x, "reactance", "x", 0.0, "fault reactance (Ohms)")
	BFCmd.Flags().StringArrayVarP(&connFlag, "conn", "c", []string{}, "fault connection codes [ABC, AG, ABG, etc.]")
	BFCmd.Flags().BoolVarP(&seqFlag, "seq", "s", false, "output in sequential components")
}

var (
	fileFlag   string
	outputFlag string
	formatFlag string
	regexFlag  string
	vminFlag   float64
	vmaxFlag   float64
	areaFlag   int
	zoneFlag   int
	r, x       float64
	connFlag   []string
	seqFlag    bool
)

var BFCmd = &cobra.Command{
	Use:     "busfault",
	Short:   "Run bus fault simulations.",
	Aliases: []string{"bf"},
	RunE:    runBusFault,
}

// result represents a bus fault simulation result.
type result struct {
	b          *model.Bus
	fd         string
	seq        bool
	va, vb, vc goolx.Phasor
	ia, ib, ic goolx.Phasor
}

func (r *result) csvHeader() []string {
	h := []string{
		"Fault Description",
		"Bus Number",
		"Bus Name",
		"Bus kV",
	}
	if r.seq {
		return append(h, []string{
			"V0_mag (kV)", "V0_ang",
			"V1_mag (kV)", "V1_ang",
			"V2_mag (kV)", "V2_ang",
			"I0_mag (A)", "I0_ang",
			"I1_mag (A)", "I1_ang",
			"I2_mag (A)", "I2_ang",
		}...)
	}
	return append(h, []string{
		"Va_mag (kV)", "Va_ang",
		"Vb_mag (kV)", "Vb_ang",
		"Vc_mag (kV)", "Vc_ang",
		"Ia_mag (A)", "Ia_ang",
		"Ib_mag (A)", "Ib_ang",
		"Ic_mag (A)", "Ic_ang",
	}...)
}

func (r *result) csvValues() []string {
	sf := fmt.Sprintf
	return []string{
		sf("%s", r.fd),
		sf("%d", r.b.Number),
		sf("%s", r.b.Name),
		sf("%0.2f", r.b.KVNominal),
		sf("%0.2f", r.va.Mag()), sf("%0.1f", r.va.Ang()),
		sf("%0.2f", r.vb.Mag()), sf("%0.1f", r.vb.Ang()),
		sf("%0.2f", r.vc.Mag()), sf("%0.1f", r.vc.Ang()),
		sf("%0.2f", r.ia.Mag()), sf("%0.1f", r.ia.Ang()),
		sf("%0.2f", r.ib.Mag()), sf("%0.1f", r.ib.Ang()),
		sf("%0.2f", r.ic.Mag()), sf("%0.1f", r.ic.Ang()),
	}
}

var results []result

// runBusFault runs bus fault simulations.
func runBusFault(cmd *cobra.Command, args []string) error {

	// Process input file flag.
	if fileFlag == "" {
		return fmt.Errorf("must provide a input file using -f")
	}
	info, err := os.Stat(fileFlag)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%q is not a file", fileFlag)
	}

	// Process fault connection flag.
	if len(connFlag) == 0 {
		return fmt.Errorf("must provide at least one fault connection using --conn, or -c flag")
	}
	var fltConns []goolx.FltConn
	for _, conn := range connFlag {
		switch strings.ToUpper(conn) {
		case "ABC":
			fltConns = append(fltConns, goolx.ABC)
		case "BCG", "CBG":
			fltConns = append(fltConns, goolx.BCG)
		case "CAG", "ACG":
			fltConns = append(fltConns, goolx.CAG)
		case "ABG", "BAG":
			fltConns = append(fltConns, goolx.ABG)
		case "AG":
			fltConns = append(fltConns, goolx.AG)
		case "BG":
			fltConns = append(fltConns, goolx.BG)
		case "CG":
			fltConns = append(fltConns, goolx.CG)
		case "BC", "CB":
			fltConns = append(fltConns, goolx.BC)
		case "CA", "AC":
			fltConns = append(fltConns, goolx.CA)
		case "AB", "BA":
			fltConns = append(fltConns, goolx.AB)
		default:
			return fmt.Errorf("unknown fault connection: %q", strings.ToUpper(conn))
		}
	}

	// Compile regular expression flag.
	var re *regexp.Regexp
	if regexFlag != "" {
		re, err = regexp.Compile(regexFlag)
		if err != nil {
			return err
		}
	}

	// Initialize OlxAPI and load case.
	api := goolx.NewClient()
	defer api.Release()

	err = api.LoadDataFile(fileFlag)
	if err != nil {
		return err
	}

	// Loop through buses.
	buses := api.NextEquipment(constants.TCBus)
	for buses.Next() {
		hnd := buses.Hnd()
		b, err := model.GetBus(api, hnd)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		// Regular expression filter.
		if re != nil {
			if !re.MatchString(b.Name) {
				// Regular expression doesn't match bus name, skipping.
				continue
			}
		}

		// Voltage limit filters.
		if b.KVNominal < vminFlag || b.KVNominal > vmaxFlag {
			continue
		}

		// Area/Zone filters.
		if areaFlag != 0 && b.Area != areaFlag {
			continue
		}
		if zoneFlag != 0 && b.Zone != zoneFlag {
			continue
		}

		// Setup fault configuration options.
		var cfgOptions = []goolx.FaultOption{
			goolx.FaultConn(fltConns...),
			goolx.FaultClearPrev(true),
			goolx.FaultCloseIn(),
		}

		if r > 0 || x > 0 {
			cfgOptions = append(cfgOptions, goolx.FaultRX(r, x))
		}

		cfg := goolx.NewFaultConfig(cfgOptions...)

		// Run bus faults.
		err = api.DoFault(hnd, cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		faults := api.NextFault(1)
		for faults.Next() {
			i := faults.Indx()
			fd := api.FaultDescription(i)
			var va, vb, vc goolx.Phasor
			switch seqFlag {
			case true:
				va, vb, vc, err = api.GetSCVoltageSeq(hnd)
			default:
				va, vb, vc, err = api.GetSCVoltagePhase(hnd)
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			var ia, ib, ic goolx.Phasor
			switch seqFlag {
			case true:
				ia, ib, ic, err = api.GetSCCurrentSeq(constants.HNDSC)
			default:
				ia, ib, ic, err = api.GetSCCurrentPhase(constants.HNDSC)
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			results = append(results, result{
				b:   b,
				fd:  fd,
				seq: seqFlag,
				va:  va, vb: vb, vc: vc,
				ia: ia, ib: ib, ic: ic,
			})
		}
	}

	var w = os.Stdout
	if outputFlag != "" {
		f, err := os.Create(outputFlag)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}

	switch {
	case strings.HasSuffix(strings.ToLower(outputFlag), ".csv") || strings.ToLower(formatFlag) == "csv":
		cw := csv.NewWriter(w)
		defer cw.Flush()
		var so sync.Once
		for _, r := range results {
			so.Do(func() {
				cw.Write(r.csvHeader())
			})
			cw.Write(r.csvValues())
		}
	default:
		tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
		defer tw.Flush()
		var osTw sync.Once
		for _, r := range results {
			osTw.Do(func() {
				tw.Write([]byte(strings.Join(r.csvHeader(), "\t") + "\n"))
			})
			tw.Write([]byte(strings.Join(r.csvValues(), "\t") + "\n"))
		}
	}

	return nil
}
