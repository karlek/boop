// Mug downloads a mugshot of KTH people with the help of xfinger.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"runtime"
	"sync"

	"github.com/karlek/boop"

	"code.google.com/p/mahonia"
	"github.com/karlek/profile"
	"github.com/mewkiz/pkg/errutil"
	"github.com/mewkiz/pkg/httputil"
)

func init() {
	flag.Usage = usage
}

func usage() {
	fmt.Fprintln(os.Stderr, os.Args[0]+" [NAMES]")
}

// Osquarulda maps a image URL and name together.
type Osquarulda struct {
	ImgUrl *url.URL
	Name   string
}

// Error wrapper.
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	if len(flag.Args()) < 1 {
		usage()
		os.Exit(0)
	}

	defer profile.Start(profile.CPUProfile).Stop()
	// Count go routines.
	osqChan := make(chan Osquarulda)
	for _, name := range flag.Args() {
		go errWrapLookup(name, osqChan)
	}

	// Download waitgroup.
	wg := new(sync.WaitGroup)
	for argc := flag.NArg(); argc > 0; argc-- {
		osq := <-osqChan
		// Ignore empty images.
		if osq.ImgUrl == nil {
			continue
		}

		go errWrapDownload(osq, wg)
		wg.Add(1)
	}
	// Wait for downloads!
	wg.Wait()
}

// errWrapLookup handles errors for lookup().
func errWrapLookup(name string, osqChan chan Osquarulda) {
	err := lookup(name, osqChan)
	if err != nil {
		log.Println(err)
		osqChan <- Osquarulda{ImgUrl: nil}
	}
}

// lookup maps a name to an image URL and sends the combination to the osqChan.
func lookup(name string, osqChan chan Osquarulda) (err error) {
	n, err := urlEncode(name)
	if err != nil {
		return errutil.Err(err)
	}
	img, err := xfinger.Single(n)
	if err != nil {
		return errutil.Err(err)
	}
	osq := Osquarulda{
		ImgUrl: img,
		Name:   name,
	}
	osqChan <- osq
	return nil
}

// urlEncode is used for xfingers ISO-8859-1 encoding of special characters.
func urlEncode(name string) (string, error) {
	n, ok := mahonia.NewEncoder("ISO-8859-1").ConvertStringOK(name)
	if !ok {
		return "", errutil.NewNoPosf("name contains non illegal charset characters: %s.", n)
	}
	return url.QueryEscape(n), nil
}

// errWrapDownload handles errors for download().
func errWrapDownload(osq Osquarulda, wg *sync.WaitGroup) {
	err := download(osq)
	if err != nil {
		log.Println(err)
	}
	wg.Done()
}

// download downloads non-placeholding Osquarulda images and saves them in the
// current folder.
func download(osq Osquarulda) (err error) {
	buf, err := httputil.Get(osq.ImgUrl.String())
	if err != nil {
		return errutil.Err(err)
	}
	// Compare the image to the placeholder image.
	if bytes.Equal(buf, defaultPic) {
		return errutil.NewNoPosf("%s missing picture.", osq.Name)
	}
	err = ioutil.WriteFile(osq.Name+".png", buf, 0777)
	if err != nil {
		return errutil.Err(err)
	}
	return nil
}

// Default picture which can be ignored.
var defaultPic = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
	0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x57,
	0x00, 0x00, 0x00, 0x69, 0x08, 0x00, 0x00, 0x00, 0x00, 0x6A, 0xD3, 0xE0,
	0x6D, 0x00, 0x00, 0x03, 0x36, 0x49, 0x44, 0x41, 0x54, 0x78, 0xDA, 0xCD,
	0xD9, 0x3D, 0xB7, 0xB3, 0x20, 0x0C, 0x00, 0xE0, 0xFE, 0x5A, 0xE7, 0xBA,
	0x5E, 0xD7, 0x76, 0x95, 0xB9, 0x73, 0x5D, 0xCB, 0xAA, 0x6B, 0x9D, 0x9D,
	0x65, 0x2D, 0xEB, 0x75, 0x95, 0x99, 0xB7, 0x5A, 0xFC, 0x42, 0x02, 0x89,
	0xF4, 0x9E, 0xF3, 0x66, 0x6C, 0xF1, 0x39, 0x34, 0x86, 0x80, 0xF6, 0xA4,
	0x83, 0xA1, 0x44, 0xC5, 0xD9, 0x35, 0xB1, 0xE2, 0xCA, 0x78, 0xA3, 0xE0,
	0x8B, 0x4E, 0x7E, 0x53, 0x56, 0xC5, 0x4E, 0x5C, 0xE3, 0x5C, 0x1E, 0x70,
	0x25, 0xCF, 0x92, 0x60, 0x64, 0x95, 0x22, 0xB9, 0x0A, 0x83, 0x7E, 0xA2,
	0x50, 0x68, 0xB7, 0xE3, 0x29, 0x56, 0x7D, 0x47, 0xCA, 0x71, 0xAE, 0x2A,
	0x08, 0xE8, 0x27, 0xD1, 0x0A, 0xE1, 0xAA, 0x2B, 0x95, 0x7D, 0x4F, 0x59,
	0x06, 0x5D, 0x89, 0x4E, 0xEC, 0x06, 0x6E, 0x02, 0xAE, 0xA4, 0x64, 0x16,
	0x9C, 0xF1, 0xCE, 0x55, 0x87, 0x66, 0xBB, 0x83, 0x6D, 0xF7, 0x48, 0x6E,
	0xA7, 0x58, 0xDF, 0x3C, 0xDB, 0x25, 0x57, 0xC2, 0x3A, 0x18, 0xE8, 0x8A,
	0x18, 0x36, 0x49, 0x2A, 0xC8, 0x8D, 0xC8, 0xC2, 0x18, 0xD2, 0xED, 0x36,
	0x91, 0x6C, 0x92, 0x29, 0xA7, 0xCB, 0x63, 0xDD, 0x84, 0x3B, 0xDD, 0xC3,
	0x35, 0xB6, 0x44, 0xE7, 0x70, 0xA5, 0xFF, 0x92, 0x73, 0xFE, 0x28, 0x3F,
	0x71, 0xBF, 0x80, 0x83, 0x98, 0xC3, 0xAD, 0x60, 0x33, 0x2F, 0xDB, 0xED,
	0x1D, 0x6E, 0x6E, 0xC0, 0x48, 0xB1, 0x77, 0xA1, 0xE2, 0xBD, 0x6F, 0xD7,
	0xBE, 0x89, 0xDF, 0x87, 0x67, 0xC2, 0x1B, 0x17, 0xAA, 0x32, 0x0D, 0x44,
	0xFB, 0x03, 0x4E, 0x78, 0xED, 0xAA, 0x84, 0xE8, 0xEA, 0xDE, 0x95, 0x67,
	0x66, 0xBB, 0x82, 0xEC, 0xBA, 0xE1, 0x8E, 0xEC, 0xB6, 0xCD, 0xBB, 0x18,
	0x9E, 0xED, 0xEF, 0x72, 0xCD, 0xCB, 0x31, 0x9C, 0x5B, 0x6E, 0xED, 0x75,
	0x5F, 0x65, 0xBE, 0x54, 0xC7, 0x52, 0x1C, 0x8E, 0x9B, 0x97, 0x59, 0x2E,
	0x87, 0xDD, 0xFE, 0x69, 0xFD, 0xE0, 0xFC, 0x35, 0x65, 0xC2, 0x31, 0x5E,
	0x62, 0xDD, 0xE7, 0x79, 0xF7, 0xD9, 0x79, 0x4A, 0x86, 0xA3, 0x8E, 0x39,
	0xD2, 0x75, 0x46, 0x6E, 0xAE, 0x7A, 0xEE, 0xBF, 0x62, 0x31, 0x6E, 0x62,
	0x26, 0xDC, 0x3A, 0x13, 0xB7, 0x76, 0x89, 0x5D, 0xB2, 0x34, 0xCB, 0xCE,
	0xF1, 0x95, 0xC0, 0xD5, 0x99, 0xD7, 0xD5, 0x8E, 0xAF, 0xF8, 0xC6, 0x95,
	0x5F, 0x73, 0x6F, 0xDB, 0xFE, 0x40, 0x3B, 0x39, 0x94, 0x70, 0x7E, 0xD9,
	0xD6, 0xA5, 0xDD, 0xB8, 0x27, 0xEC, 0x26, 0xA4, 0xBE, 0x6E, 0x85, 0xA9,
	0x87, 0x32, 0xEC, 0x92, 0xF6, 0xA1, 0xA9, 0x7E, 0x9D, 0x5B, 0x87, 0xDC,
	0xBA, 0x94, 0x53, 0x49, 0x0B, 0x97, 0xD9, 0xBB, 0xD0, 0x0E, 0xEF, 0xC7,
	0x0F, 0xED, 0x49, 0x83, 0xED, 0xE2, 0xF3, 0x70, 0xE9, 0x7D, 0xD3, 0xB5,
	0xDC, 0x9A, 0xCC, 0xEA, 0x1B, 0xC2, 0xED, 0xD0, 0xF5, 0x7B, 0xEB, 0xBD,
	0x59, 0xD8, 0xBA, 0xF8, 0x23, 0xEA, 0xB4, 0xD2, 0x5C, 0xBD, 0x6C, 0xEF,
	0x62, 0x8B, 0xE1, 0xF2, 0x0A, 0xB2, 0x6B, 0x17, 0xC9, 0x9E, 0xE7, 0xC9,
	0xEA, 0x3B, 0x3C, 0x4A, 0x9D, 0x88, 0xEC, 0x7D, 0xCA, 0xAC, 0xEE, 0x73,
	0xCF, 0xB0, 0x79, 0xBD, 0xE1, 0xD8, 0x79, 0x57, 0x83, 0xCE, 0x24, 0x26,
	0xAE, 0x93, 0x8B, 0x9C, 0xED, 0xAC, 0xF6, 0x0F, 0xEF, 0xB8, 0xA9, 0x9F,
	0x61, 0x6F, 0xD9, 0xBC, 0xB3, 0x9C, 0xFD, 0xE3, 0x4C, 0x5F, 0x47, 0xAF,
	0x07, 0xB3, 0xC6, 0xF2, 0xD0, 0xB8, 0x7A, 0x74, 0xF1, 0xCB, 0xCC, 0xBF,
	0x18, 0x96, 0x90, 0x83, 0x4B, 0xE8, 0xBA, 0x58, 0x77, 0xD8, 0x8F, 0x15,
	0x61, 0xF7, 0x41, 0xBA, 0x6C, 0x70, 0x6F, 0x18, 0x90, 0xE6, 0x0E, 0xE7,
	0x1D, 0xD2, 0xE6, 0x8E, 0x74, 0x87, 0xF3, 0x19, 0xE9, 0xC1, 0x15, 0xE7,
	0x8E, 0xE7, 0x49, 0xD2, 0xDE, 0x8E, 0x73, 0x8B, 0xC1, 0xA5, 0xB0, 0x49,
	0x3B, 0xC6, 0x3D, 0x30, 0xAA, 0x21, 0xBB, 0xA8, 0x48, 0x87, 0xDF, 0x44,
	0xCB, 0x2F, 0x2A, 0xC6, 0x47, 0xD9, 0x13, 0x7E, 0xF3, 0xC1, 0xC6, 0xF8,
	0x4C, 0x7F, 0xD2, 0xB2, 0xF8, 0xAE, 0x7C, 0xD5, 0x1F, 0x77, 0xD8, 0xD9,
	0x84, 0xA8, 0x79, 0xEC, 0xAB, 0x87, 0x29, 0xEA, 0xC5, 0x35, 0xCD, 0xEF,
	0x2B, 0x6C, 0xAA, 0x6D, 0x57, 0xB3, 0xD0, 0x35, 0x3F, 0xE6, 0x71, 0xDE,
	0x37, 0x86, 0x1F, 0x70, 0xA7, 0xA3, 0x9E, 0x6F, 0xBA, 0x6A, 0xEF, 0x06,
	0xEF, 0x1F, 0xC2, 0x2D, 0xB4, 0xA6, 0xE7, 0x37, 0xEC, 0xA6, 0x8E, 0xF7,
	0x3B, 0xE1, 0x15, 0x12, 0x76, 0xE7, 0x17, 0xB6, 0x2B, 0x37, 0x98, 0xDE,
	0xB0, 0x3B, 0x4F, 0xF7, 0xCB, 0xAE, 0xF3, 0xBD, 0x5C, 0xF8, 0xFD, 0x6C,
	0xC8, 0xCD, 0xB4, 0xCB, 0x8D, 0x7F, 0x7B, 0x26, 0x9C, 0x6E, 0x17, 0xCB,
	0x16, 0xDA, 0xE9, 0x56, 0x91, 0x6C, 0xAA, 0xDC, 0x6E, 0x6C, 0x1A, 0x6A,
	0xFD, 0x27, 0xF3, 0xCD, 0x34, 0xE0, 0x12, 0x1F, 0xBB, 0xBD, 0xD3, 0xA5,
	0xF5, 0x07, 0xFC, 0x74, 0x69, 0xFD, 0xCC, 0x17, 0x15, 0xEC, 0x46, 0xED,
	0x18, 0xF0, 0xFF, 0x22, 0x51, 0xF5, 0x6B, 0xA5, 0xE1, 0x6B, 0x75, 0x56,
	0xC0, 0x6E, 0x54, 0x1A, 0x2A, 0xD0, 0x8D, 0x5B, 0xC6, 0x02, 0x74, 0xF1,
	0x4F, 0x03, 0xAE, 0xD0, 0xA0, 0x1B, 0x95, 0xDE, 0x14, 0x76, 0x29, 0xE7,
	0xF6, 0x5D, 0x30, 0xD8, 0x8D, 0x5A, 0x15, 0x1E, 0x37, 0xEA, 0x60, 0xC9,
	0x61, 0x37, 0xAA, 0x9D, 0x79, 0xDC, 0xA8, 0x76, 0xE6, 0x71, 0xA3, 0x12,
	0xEC, 0x73, 0x63, 0xFE, 0x7B, 0x13, 0x1E, 0x57, 0xD7, 0xC7, 0x1B, 0xB0,
	0xD7, 0xD5, 0xDD, 0xE1, 0x9A, 0xF0, 0xBB, 0x43, 0x32, 0x78, 0xC1, 0x0E,
	0x34, 0xA0, 0xA0, 0x6B, 0xE2, 0xFD, 0x64, 0x20, 0x04, 0xC7, 0x2F, 0x6D,
	0xAC, 0x6B, 0x02, 0xBD, 0xB6, 0xFF, 0x13, 0x17, 0x5D, 0xD2, 0x44, 0x17,
	0x5D, 0x78, 0x34, 0x17, 0xBF, 0xB4, 0x69, 0x2E, 0xBE, 0x15, 0xD1, 0x5C,
	0x7C, 0xAB, 0xE7, 0x24, 0x97, 0xB3, 0x40, 0xA4, 0xA0, 0xFB, 0x0F, 0x46,
	0x1E, 0xB9, 0x67, 0x37, 0x72, 0xA7, 0xC6, 0x00, 0x00, 0x00, 0x3C, 0x74,
	0x45, 0x58, 0x74, 0x63, 0x6F, 0x6D, 0x6D, 0x65, 0x6E, 0x74, 0x00, 0x20,
	0x49, 0x6D, 0x61, 0x67, 0x65, 0x20, 0x67, 0x65, 0x6E, 0x65, 0x72, 0x61,
	0x74, 0x65, 0x64, 0x20, 0x62, 0x79, 0x20, 0x45, 0x53, 0x50, 0x20, 0x47,
	0x68, 0x6F, 0x73, 0x74, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x20, 0x28,
	0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x3D, 0x70, 0x6E, 0x6D, 0x72, 0x61,
	0x77, 0x29, 0x0A, 0xA3, 0x94, 0xF4, 0xC3, 0x00, 0x00, 0x00, 0x00, 0x49,
	0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82}
