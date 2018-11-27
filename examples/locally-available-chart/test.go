package main

import (
//       "bytes"
       "fmt"
//       "compress/gzip"
       "os"
//       "log"
//       "io"
       "io/ioutil"
)

func main() {

     chartTarGZName := "postgres-crd-v2-chart-0.0.2.tgz"
     chartTarFile := "./tmp1/" + chartTarGZName //+ chartName + ".tar"
     fmt.Printf("Chart Location:%s\n", chartTarFile)
     out, err := os.Create(chartTarFile)
     defer out.Close()
     if err != nil {
        //log.Fatal(err)
	panic(err)
     }

     charttgz, _ := ioutil.ReadFile(chartTarGZName)

     //var buf bytes.Buffer
     //buf.Write(charttgz)

     ioutil.WriteFile(chartTarFile, charttgz, 0644)

     /*
     zr, err := gzip.NewReader(&buf)
     if err != nil {
        //log.Fatal(err)
	panic(err)
     }

     fmt.Println("Read tgz file in buffer")
     fmt.Printf("Name: %s\nComment: %s\nModTime: %s\n\n", zr.Name, zr.Comment, zr.ModTime.UTC())

     if _, err := io.Copy(out, zr); err != nil {
        //log.Fatal(err)
	panic(err)
     }
     */
}