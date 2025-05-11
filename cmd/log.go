package cmd

import (
	"fmt"
	"math"
	"sort"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Display results",
	Long:  `Display results`,
	Run: func(cmd *cobra.Command, args []string) {
		commonInit()
		if len(args) == 0 {
			showHistory()
		} else if len(args) == 1 {
			showResults(args[0])
		}
	},
}

func showHistory() {
	fmt.Printf("%-4s %-15s %10s %10s %10s     %s\n", "No.", "Date", "GM", "AM", "Error", "Comment")
	for i := max(len(logs.vals)-30, 0); i < len(logs.vals); i++ {
		ave1, ave2, ngCnt := calcAverage(logs.vals[i])
		fmt.Printf("%04d %-15s %10d %10d %10d     %s\n", logs.idxes[i], logs.times[i], ave1, ave2, ngCnt, logs.comments[i])
	}
}
func showResults(id string) {
	var d []int
	if len(logs.vals) == 0 {
		return
	}
	d = logs.vals[len(logs.vals)-1]
	if id == "best" {
		d = logs.best
	} else {
		tgt, err := strconv.Atoi(id)
		if err != nil {
			return
		}

		ok := false
		for i := 0; i < len(logs.idxes); i++ {
			if logs.idxes[i] == tgt {
				d = logs.vals[i]
				ok = true
				break
			}
		}
		if ok == false {
			errorPrint("log not found")
			return
		}
	}

	if d == nil {
		return
	}

	type sl struct {
		ratio float64
		line  string
		score int
		rank  int
	}
	s := make([]sl, 0)
	fmt.Printf("%-4s %8s %10s     %s  %s\n", "No.", "Score", "Rank", "Parameter", "Seed")

	for i := 0; i < set.TestDataNum; i++ {
		if len(opt.filter) != 0 {
			ret := filterEvaluate(hi.HeaderData[i])
			if ret == false {
				continue
			}
		}
		green := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
		sc := 0.0
		var v string

		v = fmt.Sprintf("%10d", d[i])
		if d[i] < 0 {
			sc = -1
		}
		r := calcRank(d[i], i)
		v = green.Render(v)
		var t string
		if len(set.Seeds) != 0 {
			if set.IsSystemTest {
				t = fmt.Sprintf("%04d %s  %4d/%-4d %s   %s", i, v, r, len(logs.vals2), hi.HeaderData[i], set.Seeds[i])
			} else {
				t = fmt.Sprintf("%04d %s  %4d/%-4d %s   %s", i, v, r, len(logs.vals), hi.HeaderData[i], set.Seeds[i])
			}

		} else {
			if set.IsSystemTest {
				t = fmt.Sprintf("%04d %s  %4d/%-4d %s   %d", i, v, r, len(logs.vals2), hi.HeaderData[i], i)
			} else {
				t = fmt.Sprintf("%04d %s  %4d/%-4d %s   %d", i, v, r, len(logs.vals), hi.HeaderData[i], i)
			}

		}

		s = append(s, sl{sc, t, d[i], r})
	}
	limit := min(len(s), int(opt.linesLimit))

	if opt.order == "asc" {
		sort.Slice(s, func(i, j int) bool {
			return s[i].rank < s[j].rank
		})
	} else if opt.order == "desc" {
		sort.Slice(s, func(i, j int) bool {
			return s[i].rank > s[j].rank
		})
	}

	s1 := 0
	c1 := 0
	t1 := 0.0
	for i := 0; i < limit; i++ {
		fmt.Println(s[i].line)
		if s[i].score > 0 {
			c1++
			s1 += s[i].score
			t1 += math.Log(float64(s[i].score))
		}
	}
	if c1 != 0 {
		fmt.Println("")
		fmt.Println("[GM(AM)]")
		fmt.Printf("%d(%d)\n", int(math.Round(math.Exp(t1/float64(c1)))), s1/c1)
	}
	fmt.Println("")

}

var logClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "clear log",
	Long:  `clear log`,
	Run: func(cmd *cobra.Command, args []string) {
		commonInit()

		fmt.Println("Logs cleared")
		f := fmt.Sprintf("%s/%s", logs.logDir, RunCsv)
		renameFile(f, f+".1")
		f = fmt.Sprintf("%s/%s", logs.logDir, HistoryCsv)
		renameFile(f, f+".1")
		f = fmt.Sprintf("%s/%s", logs.logDir, ResultCsv)
		renameFile(f, f+".1")
		f = fmt.Sprintf("%s/%s", logs.logDir, InputCsv)
		renameFile(f, f+".1")
	},
}

var logDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "display diff",
	Long:  `display diff`,
	Run: func(cmd *cobra.Command, args []string) {
		commonInit()
		if len(logs.vals) < 2 {
			return
		}
		var d1, d2 []int
		var cap1, cap2 string
		if len(args) == 0 {
			d1 = logs.vals[len(logs.vals)-2]
			d2 = logs.vals[len(logs.vals)-1]
			cap1 = fmt.Sprintf("%04d", len(logs.vals)-2)
			cap2 = "last"
		} else if len(args) == 1 {
			cap2 = "last"
			d2 = logs.vals[len(logs.vals)-1]
			if args[0] == "best" {
				d1 = logs.best
				cap1 = "best"
			} else {
				tgt, err := strconv.Atoi(args[0])
				cap1 = args[0]
				if err != nil {
					return
				}

				for i := 0; i < len(logs.idxes); i++ {
					if logs.idxes[i] == tgt {
						d1 = logs.vals[i]
						break
					}
				}
			}
		} else if len(args) == 2 {
			if args[0] == "best" {
				d1 = logs.best
				cap1 = "best"
			} else {
				tgt, err := strconv.Atoi(args[0])
				if err != nil {
					return
				}
				cap1 = args[0]
				for i := 0; i < len(logs.idxes); i++ {
					if logs.idxes[i] == tgt {
						d1 = logs.vals[i]
						break
					}
				}
			}
			if args[1] == "best" {
				d2 = logs.best
				cap2 = "best"
			} else {
				tgt, err := strconv.Atoi(args[1])
				if err != nil {
					return
				}
				cap2 = args[1]
				for i := 0; i < len(logs.idxes); i++ {
					if logs.idxes[i] == tgt {
						d2 = logs.vals[i]
						break
					}
				}
			}
		}
		if d1 == nil || d2 == nil {
			return
		}

		type sl struct {
			ratio  float64
			line   string
			score1 int
			score2 int
		}
		s := make([]sl, 0)
		for i := 0; i < set.TestDataNum; i++ {
			if len(opt.filter) != 0 {
				ret := filterEvaluate(hi.HeaderData[i])
				if ret == false {
					continue
				}
			}
			diff := d1[i] - d2[i]
			blue := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true)
			red := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
			green := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
			sc := 0.0
			var v1, v2, v3, v4 string
			v1 = fmt.Sprintf("%10d", d1[i])
			v2 = fmt.Sprintf("%10d", d2[i])
			if d1[i] < 0 || d2[i] < 0 {
				sc = -1
				v3 = fmt.Sprintf("%10d", -1)
				v4 = fmt.Sprintf("%6d", -1)
				diff = 0
			} else {
				v3 = fmt.Sprintf("%10d", diff)
				v4 = fmt.Sprintf("%0.2f%%", float64(d2[i])/float64(d1[i])*100)
				sc = (float64(d2[i]) - float64(d1[i])) / float64(d1[i]) * 100
			}

			if diff > 0 {
				v1 = red.Render(v1)
				v2 = blue.Render(v2)
				v3 = blue.Render(v3)
				v4 = blue.Render(v4)
			} else if diff < 0 {
				v1 = blue.Render(v1)
				v2 = red.Render(v2)
				v3 = red.Render(v3)
				v4 = red.Render(v4)
			} else if diff == 0 {
				v1 = green.Render(v1)
				v2 = green.Render(v2)
				v3 = green.Render(v3)
				v4 = green.Render(v4)
			}

			t := fmt.Sprintf("%04d\t%s\t%s\t%s\t%s\t%s", i, hi.HeaderData[i], v1, v2, v4, v3)
			s = append(s, sl{sc, t, d1[i], d2[i]})
		}
		limit := min(len(s), int(opt.linesLimit))

		if opt.order == "asc" {
			sort.Slice(s, func(i, j int) bool {
				return s[i].ratio < s[j].ratio
			})
		} else if opt.order == "desc" {
			sort.Slice(s, func(i, j int) bool {
				return s[i].ratio > s[j].ratio
			})
		}

		s1, s2 := 0, 0
		c1, c2 := 0, 0
		t1, t2 := 0.0, 0.0
		for i := 0; i < limit; i++ {
			fmt.Println(s[i].line)
			if s[i].score1 > 0 {
				c1++
				s1 += s[i].score1
				t1 += math.Log(float64(s[i].score1))
			}
			if s[i].score2 > 0 {
				c2++
				s2 += s[i].score2
				t2 += math.Log(float64(s[i].score2))
			}
		}

		if c1 != 0 && c2 != 0 {
			fmt.Println("")
			fmt.Println("[GM(AM)]")
			fmt.Printf("%s : %d(%d)\n", cap1, int(math.Round(math.Exp(t1/float64(c1)))), s1/c1)
			fmt.Printf("%s : %d(%d)\n", cap2, int(math.Round(math.Exp(t2/float64(c2)))), s2/c2)
		}

	},
}

// init initializes the flags and subcommands for the logCmd command and its related commands
func init() {
	rootCmd.AddCommand(logCmd)
	logCmd.Flags().StringVarP(&opt.setName, "set-name", "s", "default", "Set name to run")
	logCmd.Flags().StringVarP(&opt.order, "order", "o", "", "asc or desc")
	logCmd.Flags().StringVarP(&opt.filter, "filter", "f", "", "Set filter definition")
	logCmd.AddCommand(logClearCmd)
	logClearCmd.Flags().StringVarP(&opt.setName, "set-name", "s", "default", "Set name to run")
	logCmd.AddCommand(logDiffCmd)
	logDiffCmd.Flags().StringVarP(&opt.setName, "set-name", "s", "default", "Set name to run")
	logDiffCmd.Flags().StringVarP(&opt.order, "order", "o", "", "asc or desc")
	logDiffCmd.Flags().IntVarP(&opt.linesLimit, "count", "c", INF, "max data size")
	logDiffCmd.Flags().StringVarP(&opt.filter, "filter", "f", "", "Set filter definition")
}
