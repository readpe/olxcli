// Copyright 2021 readpe All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package noclear

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/readpe/goolx"
	"github.com/readpe/goolx/model"
	"github.com/spf13/cobra"
)

func init() {
	// Subcommand Flags.
	NCCmd.Flags().StringVarP(&fileFlag, "file", "f", "", "input *.olr file")
	NCCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "output file, if ending in '.csv' will output in csv format")
	NCCmd.Flags().StringVarP(&formatFlag, "format", "F", "tab", "output file format [csv]")
	NCCmd.Flags().StringVarP(&regexFlag, "expression", "e", "", "regular expression pattern")
	NCCmd.Flags().Float64Var(&vminFlag, "vmin", 0.0, "minimum voltage (kV)")
	NCCmd.Flags().Float64Var(&vmaxFlag, "vmax", 999.0, "maximum voltage (kV)")
	NCCmd.Flags().IntVarP(&areaFlag, "area", "a", 0, "area number")
	NCCmd.Flags().IntVarP(&zoneFlag, "zone", "z", 0, "zone number")
	NCCmd.Flags().Float64VarP(&r, "resistance", "r", 0.0, "fault resistance (Ohms)")
	NCCmd.Flags().Float64VarP(&x, "reactance", "x", 0.0, "fault reactance (Ohms)")
	NCCmd.Flags().StringArrayVarP(&connFlag, "conn", "c", []string{"ABC", "AG"}, "fault connection codes [ABC, AG, ABG, etc.]")
	NCCmd.Flags().BoolVarP(&branchesFlag, "branches", "b", false, "run close-in stepped event on branches emanating from bus")
	NCCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "verbose output, show all results")
	NCCmd.Flags().Float64Var(&clearedThresholdFlag, "cleared-threshold", 1.0, "cleared threshold (amps)")
	NCCmd.Flags().Float64Var(&slowClearThresholdFlag, "slow-threshold", 3.0, "slow clear threshold (s)")
}

var (
	fileFlag               string
	outputFlag             string
	formatFlag             string
	regexFlag              string
	vminFlag               float64
	vmaxFlag               float64
	areaFlag               int
	zoneFlag               int
	r, x                   float64
	connFlag               []string
	branchesFlag           bool
	verboseFlag            bool
	clearedThresholdFlag   float64
	slowClearThresholdFlag float64
)

var NCCmd = &cobra.Command{
	Use:     "noclear",
	Short:   "Check for no or slow clearing equipment utilizing stepped event analysis",
	Aliases: []string{"nc"},
	RunE:    runNoClear,
}

var results []*result

type result struct {
	b            *model.Bus
	cfg          *goolx.SteppedEventConfig
	conn         goolx.FltConn
	cleared      bool
	currentInit  float64
	currentFinal float64
	maxTime      float64
	steps        []goolx.SteppedEvent
	note         string
}

func (r *result) csvHeader() []string {
	h := []string{
		"Fault Description",
		"Bus Number",
		"Bus Name",
		"Bus kV",
		"Max OpTime (s)",
		"Fault Cleared",
		"Initial Current (A)",
		"Final Current (A)",
		"Note",
	}
	return h
}

func (r *result) csvValues() []string {
	sf := fmt.Sprintf
	return []string{
		sf("%q", r.steps[0].FaultDescription),
		sf("%d", r.b.Number),
		sf("%s", r.b.Name),
		sf("%0.2f", r.b.KVNominal),
		sf("%0.2f", r.maxTime),
		sf("%t", r.cleared),
		sf("%0.2f", r.currentInit),
		sf("%0.2f", r.currentFinal),
		sf("%s", r.note),
	}
}

func (r *result) process() {
	for i, s := range r.steps[:len(r.steps)-1] {
		if i == 0 {
			r.currentInit = s.Current
		}
		r.currentFinal = s.Current
		r.maxTime = s.Time
		if s.Current <= clearedThresholdFlag {
			r.cleared = true
			return
		}
	}
}

// runNoClear runs no or slow clear analysis.
func runNoClear(cmd *cobra.Command, args []string) error {

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
	buses := api.NextEquipment(goolx.TCBus)
	for buses.Next() {
		busHnd := buses.Hnd()
		b, err := model.GetBus(api, busHnd)
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

		// Loop through all specified fault connections.
		for _, conn := range fltConns {

			// Setup stepped event config options.
			cfgOptions := []goolx.SteppedEventOption{
				goolx.SteppedEventConn(conn),
				goolx.SteppedEventCloseIn(),
				goolx.SteppedEventAll(),
			}
			if r > 0 || x > 0 {
				cfgOptions = append(cfgOptions, goolx.SteppedEventRX(r, x))
			}

			cfg := goolx.NewSteppedEvent(cfgOptions...)

			// Run stepped events
			err := api.DoSteppedEvent(busHnd, cfg)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}

			var res = &result{
				b:     b,
				conn:  conn,
				cfg:   cfg,
				steps: []goolx.SteppedEvent{},
			}

			nextSE := api.NextSteppedEvent()
			for nextSE.Next() {
				res.steps = append(res.steps, nextSE.Data())
			}

			results = append(results, res)

			// Don't run remaining.
			if !branchesFlag {
				continue
			}

			branches := api.NextBusEquipment(busHnd, goolx.TCBranch)
			for branches.Next() {
				brHnd := branches.Hnd()

				// Setup stepped event config options.
				cfgOptions := []goolx.SteppedEventOption{
					goolx.SteppedEventConn(conn),
					goolx.SteppedEventCloseIn(),
					goolx.SteppedEventAll(),
				}
				if r > 0 || x > 0 {
					cfgOptions = append(cfgOptions, goolx.SteppedEventRX(r, x))
				}

				cfg := goolx.NewSteppedEvent(cfgOptions...)

				// Run stepped events
				err := api.DoSteppedEvent(brHnd, cfg)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					continue
				}

				var res = &result{
					b:     b,
					conn:  conn,
					cfg:   cfg,
					steps: []goolx.SteppedEvent{},
				}

				nextSE := api.NextSteppedEvent()
				for nextSE.Next() {
					res.steps = append(res.steps, nextSE.Data())
				}

				results = append(results, res)
			}
		}
	}

	for _, r := range results {
		r.process()

		switch {
		case !r.cleared:
			r.note = "not cleared"
		case r.cleared && r.maxTime >= slowClearThresholdFlag:
			r.note = "slow clearing"
		case verboseFlag:
			r.note = "okay"
		default:
			continue
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
		tw := tabwriter.NewWriter(w, 0, 2, 1, ' ', 0)
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
