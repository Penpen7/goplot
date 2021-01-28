package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/Penpen7/goplot/cmd/energydistribution"
	"github.com/Penpen7/goplot/cmd/field"
	"github.com/Penpen7/goplot/cmd/fortbin"
	"github.com/Penpen7/goplot/cmd/phase"
	"github.com/Penpen7/goplot/cmd/physconst"
	"github.com/Penpen7/goplot/cmd/plotconfig"
	"github.com/Penpen7/goplot/cmd/simulationconfig"
	"github.com/Penpen7/goplot/cmd/utility"
)

const plotConfigFileName = "plot.json"

var plotConfig plotconfig.Art

func loadSnap(file *os.File, config simulationconfig.SimulationConfig, fileID int) bool {
	var simulationTime float32
	wg := &sync.WaitGroup{}
	start := time.Now()

	binaryReader := fortbin.ReadNextChunk(file)
	if binaryReader == nil {
		return true
	}
	binary.Read(binaryReader, binary.LittleEndian, &simulationTime)
	fmt.Println("")
	fmt.Println("読み込んでいるシミュレーション上の規格化時間:", simulationTime)

	field.LoadWriteFieldData(file, config, plotConfig, fileID, wg)
	field.LoadWriteParticleMeshData(file, config, plotConfig, fileID, wg)
	phase.LoadWritePhaseSpace(file, config, plotConfig, fileID, wg)
	energydistribution.LoadWriteEnergyDistribution(file, config, plotConfig, fileID, wg)
	fmt.Printf("\r\033[K書き込み中...")
	wg.Wait()
	fmt.Printf("\r\033[K書き込み完了\n")
	end := time.Now()
	fmt.Println("経過時間:", end.Sub(start))
	fmt.Println("")
	return false
}

func main() {
	// 時間を計測用
	start := time.Now()

	// CPU数を取得 2以上であれば並列処理される。
	fmt.Println("CPU数", runtime.NumCPU())

	plotConfig = *plotconfig.NewArt()
	plotconfig.LoadPlotConfig(&plotConfig, "plot.json")
	plotconfig.ShowPlotConfig(plotConfig)

	// outputDirectoryがあるか確認。なければディレクトリを作成する
	if err := utility.MakeDirectoryIgnoringExistance(plotConfig.OutputASCIIDirectory); err != nil {
		fmt.Printf("Error : %sが作れませんでした\n", plotConfig.OutputVTKDirectory)
		fmt.Println(err)
		os.Exit(-1)
	}
	if err := utility.MakeDirectoryIgnoringExistance(plotConfig.OutputVTKDirectory); err != nil {
		fmt.Printf("Error : %sが作れませんでした\n", plotConfig.OutputVTKDirectory)
		fmt.Println(err)
		os.Exit(-1)
	}

	// gfin.datを開き、シミュレーション設定を読み込む。
	config, err := simulationconfig.LoadSetting("gfin.dat")
	if err != nil {
		fmt.Println("gfin.datが読み込めません")
		fmt.Println(err)
		os.Exit(-1)
	}
	simulationconfig.ShowConfig(config)

	// snapのバイナリを開く(とりあえずここではsnap0001.dat)
	file, err := os.Open("snap0001.dat")
	if err != nil {
		fmt.Println("snap0001.datが読み込めません")
		fmt.Println(err)
		os.Exit(-1)
	}

	// 設定を表示する
	fmt.Println("")
	fmt.Println("シミュレーションの設定")
	physconst.CalculateNormalizeConstant(config)

	// snapを終端に達するまで読み込む。
	for fileID := 0; ; fileID++ {
		if loadSnap(file, config, fileID) {
			break
		}
	}

	// 終了時間を記憶
	end := time.Now()
	fmt.Println("正常終了")

	// 終了時間と開始時間の差をとる=かかった時間
	fmt.Println("総時間", end.Sub(start))
}
