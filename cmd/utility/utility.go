package utility

import (
	"os"
)

// 1次元配列を3次元配列に変換します。
func Slice1Dto3D(slice1D []float32, xsize int32, ysize int32, zsize int32, normalizeConstant float32) [][][]float32 {
	// 3次元スライス作成
	var slice3D [][][]float32
	slice3D = make([][][]float32, xsize)
	for x := int32(0); x < xsize; x++ {
		slice3D[x] = make([][]float32, ysize)
		for y := int32(0); y < ysize; y++ {
			slice3D[x][y] = make([]float32, zsize)
		}
	}

	// 1次元配列から、3次元配列に割り当てる。
	index2D := 0
	for y := int32(0); y < ysize; y++ {
		for z := int32(0); z < zsize; z++ {
			for x := int32(0); x < xsize; x++ {
				slice3D[x][y][z] = slice1D[index2D] * normalizeConstant
				index2D++
			}
		}
	}

	return slice3D
}

// y-px, y-py, y-pzで出力される1次元配列を2次元配列に整形します。
func Transpy(yp []float32, Ny_d int, parallelNumber int, momentumMeshNumber int) [][]float32 {
	var res [][]float32 = make([][]float32, Ny_d)
	for y := 0; y < Ny_d; y++ {
		res[y] = make([]float32, len(yp)/int(Ny_d))
	}

	Ny_d_pe := Ny_d / parallelNumber

	for ypindex, v := range yp {
		mype := ypindex / (Ny_d_pe * momentumMeshNumber)
		ipe := ypindex - mype*Ny_d_pe*momentumMeshNumber
		kp := (ipe) / Ny_d_pe
		ky_pe := (ipe) % (Ny_d_pe)
		kynew := Ny_d_pe*mype + ky_pe
		res[kynew][kp] = v
	}
	return res
}

// 1次元配列を2次元配列に変換します。
func Slice1Dto2D(g []float32, xsize int32, ysize int32) [][]float32 {
	var g2D [][]float32
	g2D = make([][]float32, xsize)
	for x := int32(0); x < xsize; x++ {
		g2D[x] = make([]float32, ysize)
	}
	index := 0
	for y := int32(0); y < ysize; y++ {
		for x := int32(0); x < xsize; x++ {
			g2D[x][y] = g[index]
			index++
		}
	}
	return g2D
}

// 存在を無視して、ディレクトリを作成します。
func MakeDirectoryIgnoringExistance(dname string) error {
	if _, err := os.Stat(dname); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(dname, 0777); err != nil {
				return err
			}
		}
	}
	return nil
}
