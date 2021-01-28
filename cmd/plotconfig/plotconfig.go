package plotconfig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Subart struct {
	Name   string
	Plot   bool
	Center string
}
type Art struct {
	OutputASCIIDirectory string
	OutputVTKDirectory   string
	Field                []Subart
	Particle             []Subart
	Phase                []Subart
	EnergyDistribution   []Subart
}

func LoadPlotConfig(v *Art, plotConfigFileName string) {
	if _, err := os.Stat(plotConfigFileName); os.IsNotExist(err) {
		fmt.Printf("\x1b[35mwarning : %sが存在しないため、新規作成しました。プロットしたいデータを変える場合、%sを変更してください。\n", plotConfigFileName, plotConfigFileName)
		fmt.Printf("Name:データの名前\n")
		fmt.Printf("Plot:出力するか否か\n")
		fmt.Printf("Center:どのデータをプロットするか(複数ある場合はスペース区切りで指定)\x1b[0m\n")

		file, _ := os.Create(plotConfigFileName)
		var buf2 bytes.Buffer
		buf, _ := json.Marshal(*NewArt())
		json.Indent(&buf2, buf, "", "  ")
		file.Write(buf2.Bytes())
		file.Close()
	}
	buf, err := ioutil.ReadFile(plotConfigFileName)

	err = json.Unmarshal(buf, v)
	if err != nil {
		fmt.Println(err)
	}
	return
}
func NewArt() *Art {
	var tempart Art
	tempart.Field = append(tempart.Field, Subart{"Ex", true, "xy x y"})
	tempart.Field = append(tempart.Field, Subart{"Ey", true, "xy x y"})
	tempart.Field = append(tempart.Field, Subart{"Ez", false, "xy x y"})
	tempart.Field = append(tempart.Field, Subart{"Bx", false, "xy x y"})
	tempart.Field = append(tempart.Field, Subart{"By", false, "xy x y"})
	tempart.Field = append(tempart.Field, Subart{"Bz", true, "xy x y"})
	tempart.Field = append(tempart.Field, Subart{"Jx", true, "xy x y"})
	tempart.Field = append(tempart.Field, Subart{"Jy", true, "xy x y"})
	tempart.Field = append(tempart.Field, Subart{"Jz", false, "xy x y"})
	tempart.Particle = append(tempart.Particle, Subart{"Ion_Density", true, "xy x y"})
	tempart.Particle = append(tempart.Particle, Subart{"Ion_Energy", false, "xy x y"})
	tempart.Particle = append(tempart.Particle, Subart{"Ion_Energy_Distribution", true, ""})
	tempart.Particle = append(tempart.Particle, Subart{"Ion_Energy_DistributionLogLog", true, ""})
	tempart.Particle = append(tempart.Particle, Subart{"Ion_EnergyFlux_x", false, "xy x y"})
	tempart.Particle = append(tempart.Particle, Subart{"Ion_EnergyFlux_y", false, "xy x y"})
	tempart.Particle = append(tempart.Particle, Subart{"Electron_Density", true, "xy x y"})
	tempart.Particle = append(tempart.Particle, Subart{"Electron_Energy", false, "xy x y"})
	tempart.Particle = append(tempart.Particle, Subart{"Electron_Energy_Distribution", true, ""})
	tempart.Particle = append(tempart.Particle, Subart{"Electron_Energy_DistributionLogLog", true, ""})
	tempart.Particle = append(tempart.Particle, Subart{"Electron_EnergyFlux_x", false, "xy x y"})
	tempart.Particle = append(tempart.Particle, Subart{"Electron_EnergyFlux_y", false, "xy x y"})
	tempart.OutputASCIIDirectory = "biny_dataASCII"
	tempart.OutputVTKDirectory = "biny_dataVTK"
	return &tempart
}
func SearchSubart(subart []Subart, name string) bool {
	for _, v := range subart {
		if v.Name == name {
			return true
		}
	}
	return false
}
func ShowPlotConfig(config Art) {
	fmt.Println("")
	fmt.Printf("出力先のディレクトリ(テキストファイル) : %s\n", config.OutputASCIIDirectory)
	fmt.Printf("出力先のディレクトリ(VTKファイル))     : %s\n", config.OutputVTKDirectory)
	fmt.Println("")
	fmt.Println("出力するデータ")
	for _, v := range config.Field {
		if v.Plot {
			fmt.Printf("%s : %s\n", v.Name, strings.Replace(v.Center, " ", ", ", -1))
		}
	}
	for _, v := range config.Particle {
		if v.Plot {
			fmt.Printf("%s : %s\n", v.Name, strings.Replace(v.Center, " ", ", ", -1))
		}
	}
}
