// package exif

// import (
// 	"fmt"
// 	"log"
// 	"os"

// 	"github.com/rwcarlsen/goexif/exif"
// 	"github.com/rwcarlsen/goexif/tiff"
// )

// type exifPrinter struct{}

// func (p exifPrinter) Walk(name exif.FieldName, tag *tiff.Tag) error {
// 	val, err := tag.StringVal()
// 	if err != nil {
// 		val = tag.String()
// 	}
// 	fmt.Printf("%-30s: %s\n", name, val)
// 	return nil
// }

// func AnalyzeEXIF(filePath string) {
// 	f, err := os.Open(filePath)
// 	if err != nil {
// 		log.Fatalf("Fehler beim Öffnen der Datei: %v", err)
// 	}
// 	defer f.Close()

// 	x, err := exif.Decode(f)
// 	if err != nil {
// 		fmt.Println("Keine EXIF-Daten gefunden – verdächtig!")
// 		return
// 	}

// 	fmt.Println("  Allgemeine EXIF-Daten:\n")
// 	x.Walk(exifPrinter{})

// 	// Einzelne wichtige Felder
// 	fmt.Println("\n📌 Einzelne Werte:")

// 	if dt, err := x.DateTime(); err == nil {
// 		fmt.Println("Aufnahmedatum:", dt)
// 	}

// 	if tag, err := x.Get(exif.Make); err == nil {
// 		fmt.Println("Kamerahersteller:", tag.String())
// 	}

// 	if tag, err := x.Get(exif.Model); err == nil {
// 		fmt.Println("Kameramodell:", tag.String())
// 	}

// 	if tag, err := x.Get(exif.LensModel); err == nil {
// 		fmt.Println("Objektiv:", tag.String())
// 	}

// 	if sw, err := x.Get(exif.Software); err == nil {
// 		val, _ := sw.StringVal()
// 		fmt.Println("Bearbeitungssoftware:", val)
// 		if val == "Adobe Photoshop" {
// 			fmt.Println("⚠️  Hinweis: Bild wurde bearbeitet")
// 		}
// 	}

// 	// GPS-Koordinaten
// 	fmt.Println("\n  GPS-Daten:")
// 	if lat, long, err := x.LatLong(); err == nil {
// 		fmt.Printf("GPS-Koordinaten: %.6f, %.6f\n", lat, long)
// 	} else {
// 		fmt.Println("Keine GPS-Koordinaten vorhanden.")
// 	}

// 	// Alle Tags in allen IFDs
// 	fmt.Println("\n Rohe EXIF-Tags (alle IFDs):")
// 	rawExif := x.Raw
// 	fmt.Println(rawExif)

// }
// pkg/analyzer/exif/exif_analyzer.go
package exifanalyzer

import (
	"os"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

type EXIFData struct {
	DateTime *time.Time  `json:"date_time,omitempty"`
	Make     string      `json:"make,omitempty"`
	Model    string      `json:"model,omitempty"`
	GPS      *[2]float64 `json:"gps,omitempty"`
	Raw      []byte      `json:"raw,omitempty"`
}

func AnalyzeEXIF(path string) (*EXIFData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		// kein EXIF im Bild ⇒ gib ein leeres []byte zurück
		return &EXIFData{Raw: []byte{}}, nil
	}

	out := &EXIFData{
		Raw: x.Raw,
	}

	if dt, err := x.DateTime(); err == nil {
		out.DateTime = &dt
	}
	if lat, lon, err := x.LatLong(); err == nil {
		out.GPS = &[2]float64{lat, lon}
	}
	if tag, err := x.Get(exif.Make); err == nil {
		out.Make, _ = tag.StringVal()
	}
	if tag, err := x.Get(exif.Model); err == nil {
		out.Model, _ = tag.StringVal()
	}

	return out, nil
}
