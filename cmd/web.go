package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

// webCmd represents the web command
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Display standings or the histogram of parameters",
	Long:  `Display standings or the histogram of parameters`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// webStandingsCmd represents the standings subcommand of web
var webStandingsCmd = &cobra.Command{
	Use:   "standings",
	Short: "Displays the ranking table of a test set.",
	Long:  `Displays the ranking table of a test set.`,
	Run: func(cmd *cobra.Command, args []string) {
		commonInit()
		showStandings()
	},
}

// webParamCmd represents the param subcommand of web
var webParamCmd = &cobra.Command{
	Use:   "param",
	Short: "Displays a histogram for each parameter of a test set.",
	Long:  `Displays a histogram for each parameter of a test set.`,
	Run: func(cmd *cobra.Command, args []string) {
		commonInit()
		showHistogram()
	},
}

func noCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		next.ServeHTTP(w, r)
	})
}
func showStandings() {
	t := fmt.Sprintf("%s", logs.logRootDir)
	if !changeDir(t) {
		return
	}
	go func() {
		//サーバーが起動するのを少し待つ
		var contestType string
		if conf.Common.IsRankMin == true {
			contestType = "rank_min"
		} else {
			contestType = "rank_max"
		}
		url := fmt.Sprintf("http://localhost:8080/?contest=%s&contest_type=%s", set.SetName, contestType)
		time.Sleep(1 * time.Second)
		// ブラウザを開く
		err := open.Start(url)
		if err != nil {
			log.Fatal("Failed to open browser: ", err)
		}
	}()

	// httpサーバーを起動する
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", noCache(fs))
	log.Println("Serving on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
		return
	}
}
func showHistogram() {
	go func() {
		//サーバーが起動するのを少し待つ
		time.Sleep(1 * time.Second)
		// ブラウザを開く
		err := open.Start("http://localhost:8080")
		if err != nil {
			log.Fatal("Failed to open browser: ", err)
		}
	}()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// ヒストグラムデータをJSON形式に変換
		dataJson, err := json.Marshal(getHistgramData())
		if err != nil {
			log.Fatalf("JSON Marshal error: %v", err)
		}
		// テンプレートにデータを渡してレンダリング
		err = tmpl.Execute(w, map[string]interface{}{
			"Title":      fmt.Sprintf("%s", cmn.ContestName),
			"DataHeader": hi.Header,
			"DataJson":   template.JS(dataJson),
		})
		if err != nil {
			log.Fatalf("Template execution error: %v", err)
		}
	})
	log.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func getHistgramData() map[string][]float64 {
	ret := make(map[string][]float64)
	for i := 0; i < len(hi.HeaderData); i++ {
		for j := 0; j < len(hi.Header); j++ {
			v, _ := strconv.ParseFloat(hi.HeaderData[i][j], 64)
			ret[hi.Header[j]] = append(ret[hi.Header[j]], v)
		}
	}
	return ret
}

// ヒストグラム表示用のHTMLテンプレート
var tmpl = template.Must(template.New("histogram").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}}</title>
    <script src="https://cdn.plot.ly/plotly-latest.min.js"></script>
</head>
<body>

<!-- データの種類を選択するドロップダウンメニュー -->
<select id="dataSelector" onchange="updatePlot()">
    {{range $_, $key := .DataHeader}}
    <option value="{{$key}}">パラメータ {{$key}}</option>
    {{end}}
</select>

<div id="plot"></div>

<script>
var allData = {{.DataJson}};

function updatePlot() {
    var selectedType = document.getElementById('dataSelector').value;
    var data = allData[selectedType];

    var trace = {
        x: data,
        type: 'histogram',
    };

    var layout = {
        title: {{.Title}},
        xaxis: {title: 'Value'},
        yaxis: {title: 'Count'}
    };

    Plotly.newPlot('plot', [trace], layout);
}

// 初期描画
updatePlot();
</script>

</body>
</html>
`))

func init() {
	rootCmd.AddCommand(webCmd)
	webCmd.AddCommand(webStandingsCmd)
	webCmd.AddCommand(webParamCmd)
	webStandingsCmd.Flags().StringVarP(&opt.setName, "set-name", "s", "default", "Set name to run")
	webParamCmd.Flags().StringVarP(&opt.setName, "set-name", "s", "default", "Set name to run")
}
