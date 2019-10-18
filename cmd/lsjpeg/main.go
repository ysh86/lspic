package main

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"os"

	"github.com/ysh86/lsjpeg"
)

func main() {
	// args
	var (
		srcFile string
	)
	if len(os.Args) > 1 && os.Args[1] != "-h" {
		srcFile = os.Args[1]
	} else {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "  string")
		fmt.Fprintln(os.Stderr, "\tsrc file")
		return
	}

	file, err := os.Open(srcFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		panic(err)
	}

	jpegFile, err := lsjpeg.NewFile(io.NewSectionReader(file, 0, stat.Size()))
	if err != nil {
		panic(err)
	}
	if err := jpegFile.Parse(); err != nil {
		panic(err)
	}

	fxmp, err := os.Create(srcFile + ".xmp")
	if err != nil {
		panic(err)
	}
	defer fxmp.Close()

	for _, s := range jpegFile.Segments {
		if err := s.Parse(); err != nil {
			panic(err)
		}
		s.DumpTo(os.Stdout, fxmp)
	}

	// XML:
	/*
		<x:xmpmeta xmlns:x="adobe:ns:meta/" x:xmptk="Adobe XMP Core 5.1.0-jc003">
		<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
		  <rdf:Description rdf:about=""
			  xmlns:GFocus="http://ns.google.com/photos/1.0/focus/"
			  xmlns:GImage="http://ns.google.com/photos/1.0/image/"
			  xmlns:GDepth="http://ns.google.com/photos/1.0/depthmap/"
			  xmlns:GCamera="http://ns.google.com/photos/1.0/camera/"
			  xmlns:xmpNote="http://ns.adobe.com/xmp/note/"
			GFocus:BlurAtInfinity="0.028364485"
			GFocus:FocalDistance="12.215591"
			GFocus:DepthOfField="0.1"
			GFocus:FocalPointX="0.5"
			GFocus:FocalPointY="0.5"
			GImage:Mime="image/jpeg"
			GDepth:Format="RangeInverse"
			GDepth:Near="8.538637161254883"
			GDepth:Far="61.86031723022461"
			GDepth:Mime="image/jpeg"
			xmpNote:HasExtendedXMP="623F9F3BD062B9F6D7A508F62F69908D">
			<GCamera:SpecialTypeID>
			  <rdf:Bag>
				<rdf:li>com.google.android.apps.camera.gallery.specialtype.SpecialType-REFOCUS</rdf:li>
			  </rdf:Bag>
			</GCamera:SpecialTypeID>
		  </rdf:Description>
		</rdf:RDF>
		</x:xmpmeta>

		<x:xmpmeta xmlns:x="adobe:ns:meta/" x:xmptk="Adobe XMP Core 5.1.0-jc003">
		<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
		  <rdf:Description rdf:about=""
			  xmlns:GImage="http://ns.google.com/photos/1.0/image/"
			  xmlns:GDepth="http://ns.google.com/photos/1.0/depthmap/"
			GImage:Data="base64..."
			GDepth:Data="base64..."/>
		</rdf:RDF>
		</x:xmpmeta>
	*/

	/*
		<?xpacket begin="<U+FEFF>" id="W5M0MpCehiHzreSzNTczkc9d"?>
		<x:xmpmeta xmlns:x="adobe:ns:meta/" x:xmptk="Adobe XMP Core 5.0-c060 61.134777, 2010/02/12-17:32:00        ">
		<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
		  <rdf:Description rdf:about=""
			  xmlns:GDepth="http://ns.google.com/photos/1.0/depthmap/"
			  xmlns:GImage="http://ns.google.com/photos/1.0/image/"
			  xmlns:GCamera="http://ns.google.com/photos/1.0/camera/"
			  xmlns:GCreations="http://ns.google.com/photos/1.0/creations/"
			  xmlns:xmp="http://ns.adobe.com/xap/1.0/"
			  xmlns:photoshop="http://ns.adobe.com/photoshop/1.0/"
			  xmlns:dc="http://purl.org/dc/elements/1.1/"
			  xmlns:xmpMM="http://ns.adobe.com/xap/1.0/mm/"
			  xmlns:stEvt="http://ns.adobe.com/xap/1.0/sType/ResourceEvent#"
			  xmlns:xmpNote="http://ns.adobe.com/xmp/note/"
			  GDepth:Mime="image/jpeg"
			  GDepth:Format="RangeInverse"
			  GDepth:Near="0.166053"
			  GDepth:Far="0.680994"
			  GImage:Mime="image/jpeg"
			  GCamera:PortraitNote=""
			  GCamera:PortraitRequest="base64..."
			  GCamera:PortraitVersion="0"
			  GCamera:BurstID="8bafacec-2d9c-415a-8cea-91a9c041186f"
			  GCamera:BurstPrimary="1"
			  GCreations:CameraBurstID="8bafacec-2d9c-415a-8cea-91a9c041186f"
			  xmp:ModifyDate="2018-10-17T14:35:05+09:00"
			  xmp:CreateDate="2018-10-16T13:11:37.349430+09:00"
			  xmp:CreatorTool="HDR+ 1.0.215421313z"
			  xmp:MetadataDate="2018-10-17T14:35:05+09:00"
			  photoshop:ColorMode="3"
			  photoshop:ICCProfile="sRGB IEC61966-2.1"
			  photoshop:DateCreated="2018-10-16T13:11:38.537649024"
			  dc:format="image/jpeg"
			  xmpMM:InstanceID="xmp.iid:0AEB6999CBD1E811A209C075810D8F37"
			  xmpMM:DocumentID="xmp.did:0AEB6999CBD1E811A209C075810D8F37"
			  xmpMM:OriginalDocumentID="xmp.did:0AEB6999CBD1E811A209C075810D8F37"
			  xmpNote:HasExtendedXMP="44FCD32E64C9269B059A108E2F054713">
			  <GCamera:SpecialTypeID>
				<rdf:Bag>
					<rdf:li>com.google.android.apps.camera.gallery.specialtype.SpecialType-PORTRAIT</rdf:li>
				</rdf:Bag>
			  </GCamera:SpecialTypeID>
			  <xmpMM:History>
				<rdf:Seq>
				<rdf:li stEvt:action="saved" stEvt:instanceID="xmp.iid:0AEB6999CBD1E811A209C075810D8F37" stEvt:when="2018-10-17T14:35:05+09:00" stEvt:softwareAgent="Adobe Photoshop CS5 Windows" stEvt:changed="/"/>
				</rdf:Seq>
			  </xmpMM:History>
		  </rdf:Description>
		</rdf:RDF>
		</x:xmpmeta>
		<?xpacket end="w"?>

		<x:xmpmeta xmlns:x="adobe:ns:meta/" x:xmptk="Adobe XMP Core 5.0-c060 61.134777, 2010/02/12-17:32:00        ">
		<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
		  <rdf:Description rdf:about=""
			  xmlns:crs="http://ns.adobe.com/camera-raw-settings/1.0/"
			  xmlns:GImage="http://ns.google.com/photos/1.0/image/"
			  xmlns:GDepth="http://ns.google.com/photos/1.0/depthmap/"
			  crs:AlreadyApplied="True"
			GImage:Data="base64..."
			GDepth:Data="base64..."/>
		</rdf:RDF>
		</x:xmpmeta>
	*/

	type XMP struct {
		XMLName xml.Name `xml:"adobe:ns:meta/ xmpmeta"`

		RDF struct {
			Description struct {
				DepthMime   string  `xml:"http://ns.google.com/photos/1.0/depthmap/ Mime,attr"`
				DepthFormat string  `xml:"http://ns.google.com/photos/1.0/depthmap/ Format,attr"`
				DepthNear   float64 `xml:"http://ns.google.com/photos/1.0/depthmap/ Near,attr"`
				DepthFar    float64 `xml:"http://ns.google.com/photos/1.0/depthmap/ Far,attr"`
				DepthData   string  `xml:"http://ns.google.com/photos/1.0/depthmap/ Data,attr"`
				ImageMime   string  `xml:"http://ns.google.com/photos/1.0/image/ Mime,attr"`
				ImageData   string  `xml:"http://ns.google.com/photos/1.0/image/ Data,attr"`
			} `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# Description"`
		} `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# RDF"`
	}

	_, err = fxmp.Seek(0, io.SeekStart)
	if err != nil {
		panic(err)
	}

	xmps := make([]*XMP, 0, 2)
	dec := xml.NewDecoder(fxmp)
	for {
		xmp := &XMP{}
		err := dec.Decode(xmp)
		if err != nil {
			break
		}
		xmps = append(xmps, xmp)
	}

	// Validate
	if len(xmps) != 2 ||
		len(xmps[1].RDF.Description.DepthData) == 0 ||
		len(xmps[1].RDF.Description.ImageData) == 0 ||
		(xmps[0].RDF.Description.DepthMime != "image/jpeg" &&
		xmps[1].RDF.Description.DepthMime != "image/jpeg" &&
		xmps[0].RDF.Description.DepthMime != "image/png" &&
		xmps[1].RDF.Description.DepthMime != "image/png") ||
		(xmps[0].RDF.Description.ImageMime != "image/jpeg" && xmps[1].RDF.Description.ImageMime != "image/jpeg") {
		panic(fmt.Errorf("Unknown XMP format"))
	}

	// merge
	xmp := xmps[0]
	if xmp.RDF.Description.DepthMime == "" {
		xmp.RDF.Description.DepthMime = xmps[1].RDF.Description.DepthMime
	}
	if xmp.RDF.Description.ImageMime == "" {
		xmp.RDF.Description.ImageMime = xmps[1].RDF.Description.ImageMime
	}
	if xmp.RDF.Description.DepthFormat == "" {
		xmp.RDF.Description.DepthFormat = xmps[1].RDF.Description.DepthFormat
	}
	if xmp.RDF.Description.DepthNear == 0.0 {
		xmp.RDF.Description.DepthNear = xmps[1].RDF.Description.DepthNear
	}
	if xmp.RDF.Description.DepthFar == 0.0 {
		xmp.RDF.Description.DepthFar = xmps[1].RDF.Description.DepthFar
	}
	xmp.RDF.Description.DepthData = xmps[1].RDF.Description.DepthData
	xmp.RDF.Description.ImageData = xmps[1].RDF.Description.ImageData

	depthData, err := base64.StdEncoding.DecodeString(xmp.RDF.Description.DepthData)
	if err != nil {
		panic(err)
	}
	imageData, err := base64.StdEncoding.DecodeString(xmp.RDF.Description.ImageData)
	if err != nil {
		panic(err)
	}

	// dump
	var depthName string
	if xmp.RDF.Description.DepthMime == "image/jpeg" {
		depthName = srcFile + ".depth.jpg"
	} else {
		depthName = srcFile + ".depth.png"
	}
	fdepth, err := os.Create(depthName)
	if err != nil {
		panic(err)
	}
	defer fdepth.Close()
	_, err = fdepth.Write(depthData)
	if err != nil {
		panic(err)
	}

	fimage, err := os.Create(srcFile + ".image.jpg")
	if err != nil {
		panic(err)
	}
	defer fimage.Close()
	_, err = fimage.Write(imageData)
	if err != nil {
		panic(err)
	}

	fmt.Printf("xmp: depth format=%s, near=%f, far=%f\n",
		xmp.RDF.Description.DepthFormat,
		xmp.RDF.Description.DepthNear,
		xmp.RDF.Description.DepthFar,
	)
}
