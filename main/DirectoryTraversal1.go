package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"fmt"
	"time"
	"sync"
)

var wg sync.WaitGroup
func main() {
	start:=time.Now().Nanosecond()
	flag.Parse()
	roots:=flag.Args()

	var defaultDir = "C:\apache-ftpserver-1.0.5"
	fileSizeChan := make(chan int64)

	//var wg sync.WaitGroup{}

	if len(roots)==0{
		roots = []string{defaultDir}
	}

	var fileCount, nbytes int64

	go func(){
		for _,dir:= range roots {
			wg.Add(1)
			go walkDir(dir, fileSizeChan)
		}
		//close(fileSizeChan)
	}()

	go func() {
		wg.Wait()
		close(fileSizeChan)
	}()

	/*for size:= range fileSizeChan{
		fileCount++
		nbytes+=size
	}*/

	tick:=time.Tick(1*time.Second)
	outerloop:
	for{
		select {
		case <- tick: //possible goroutine leak!!!
			printDiskUsage(fileCount, nbytes)
		case size, isEP := <-fileSizeChan:
			if (!isEP){
				break outerloop
			} else{
				fileCount++
				nbytes+=size
			}
		}
	}


	printDiskUsage(fileCount, nbytes)


	fmt.Println("total time taken: ", (time.Now().Nanosecond()-start))
}
func walkDir(dir string, fileSizeChan chan <- int64){
	defer wg.Done()
	for  _,entry := range directoryEntries(dir){
		if entry.IsDir(){
			subdir:=filepath.Join(dir, entry.Name())
			wg.Add(1)
			go walkDir(subdir, fileSizeChan)
		} else{
			fileSizeChan <- entry.Size()
		}
	}
}
var sem chan int = make (chan int, 2)
func directoryEntries(dir string) ([]os.FileInfo){

	sem <- 1
	defer func (){
	    <-sem
	}()

	retVal, err := ioutil.ReadDir(dir)
	if err == nil{
		return retVal
	}else{
		return nil
	}
}

func printDiskUsage(nfiles, nbytes int64) {
	fmt.Printf("%d files %.1f GB\n", nfiles, float64(nbytes)/1e9)
}