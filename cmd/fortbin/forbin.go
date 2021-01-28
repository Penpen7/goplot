package fortbin

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// ヘッダとフッタが含まれるバイナリファイルを読み込みます。
func ReadNextChunk(file *os.File) *bytes.Buffer {
	// TODO:読み込みを並行化させた方がいい
	const HEADERSIZE = 4
	const FOOTERSIZE = 4
	l := make([]byte, HEADERSIZE)

	// seek 4byte
	file.Read(l)
	var size int32
	binary.Read(bytes.NewBuffer(l), binary.LittleEndian, &size)

	// seek size byte
	m := make([]byte, size)
	binary.Read(file, binary.LittleEndian, &m)
	// fmt.Println(size)
	// seek 4byte
	n := make([]byte, FOOTERSIZE)
	_, err := file.Read(n)
	if err == io.EOF {
		fmt.Println("ファイルの終端に達しました")
		return nil
	} else if err != nil {
		panic(err)
	}
	return bytes.NewBuffer(m)
}
