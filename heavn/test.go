package main

import (
    "flag"
    "fmt"
    //~ "io"
    "io/ioutil"
    "os"
    "math"
    "bufio"
    "strconv"
    
    "image"
    _ "image/color"
    _ "image/gif"
    _ "image/jpeg"
    _ "image/png"
    
)

var persons map[string]*Histogram


func decode(filename string) (image.Image, string, error) {
    f, err := os.Open(filename)
    if err != nil {
        return nil, "no image", err
    }
    defer f.Close()
    rd := bufio.NewReader(f)
    return image.Decode(rd)
}

func handle(m image.Image, r image.Rectangle) (mat *Matrix) {
    i := 0
    //~ mat = createMatrix(r.Max.X-r.Min.X,r.Max.Y-r.Max.Y)
    mat = createMatrix(r.Max.X,r.Max.Y)
    for y := r.Min.Y; y < r.Max.Y; y++ {
        for x := r.Min.X; x < r.Max.X; x++ {
            r8,g8,b8,_ := m.At(x,y).RGBA()
            mat.e[i] = uint8((r8>>2) + (g8>>1) + (b8>>2))
            i += 1
        }
    }
    return
}
func print(h *Histogram) {
    n := 0
    for k, v := range *h {
        if v > 0 {
            fmt.Printf("%3v %2v; ", k, v)
            n += 1
        }
    }
    fmt.Printf("(%d)\n", n)
}


func compare(match string, hist *Histogram) (best string, confidence float64) {
    if hist==nil { return }
    if len(persons)<2 { return }
    best = ""
    bd := 10000000.0
    mm := 0.00000001
    for n,h := range persons {
        if match == n { continue }
        dist := norml2(h,hist)
        if dist < bd {
            bd = dist
            best = n
        }
        mm = math.Max(mm,dist)
    }
    confidence = 1.0-(bd/mm)
    return 
}



func main() {
    initIndices()
    persons = make(map[string]*Histogram)
    path := "E:/MEDIA/faces/tv"


    flag.Parse()
    if flag.NArg() > 0 {
        path = flag.Arg(0)
    }

    an := "square"
    if flag.NArg() > 1 {
        an = flag.Arg(1)
    }

    var sample Sampler = square
    if an == "square" { sample = square  }
    if an == "circle" { sample = circle  }
    if an == "circ2"  { sample = circle2 }
    if an == "sqr2"   { sample = square2 }
    if an == "elbp"   { sample = elbp    }


    var radius int = 2
    if flag.NArg() > 2 {
        radius,_ = strconv.Atoi(flag.Arg(2))
    }

    inf, err := ioutil.ReadDir(path) 
    if err != nil {
        fmt.Println("Dir:", err)
        return 
    }
    
    //~ fmt.Println("train",len(inf)/2)
    for i,d := range(inf) {
        if i%2==0 { continue }
        im, _, err := decode(path + "/" + d.Name())
        if err != nil {
            fmt.Println("Train:", err)
            continue
        }
        m := handle(im, im.Bounds())
        hist := m.histogram(sample, radius)
        persons[d.Name()] = hist
        if i > 600 { break }
    }

    //~ fmt.Println("test",len(inf)/2)
    misses := 0
    meanc := 0.0
    bestc := 0.0
    bestm := 0.0
    wostc := 100000000.0
    for i,d := range(inf) {
        if i%2!=0 { continue }
        im, _, err := decode(path + "/" + d.Name())
        if err != nil {
            fmt.Println("Test:", err)
            continue
        }
        m := handle(im, im.Bounds())
        hist := m.histogram(sample, radius)
        persons[d.Name()] = hist
        best,confidence := compare(d.Name(),hist)
        hit := (best[:4] == d.Name()[:4])
        meanc += confidence
        bestc = math.Max(confidence,bestc)
        if !hit { 
            bestm = math.Max(confidence,bestm)
            misses++ 
        } else {
            wostc = math.Min(confidence,wostc)
        }
        //~ fmt.Printf("Decode:%5v %20v %20v %3.4f\n", hit, d.Name(),best,confidence)
        if i > 600 { break }
    }
    meanc /= float64(len(inf)/2)
    fmt.Printf("%6s_%d: %3d/%-4d(%3.3f) misses %3.3f %3.3f %3.3f %3.3f %4.3f\n", an,radius, misses,len(inf)/2,float64(misses)/float64(len(inf)/2),meanc,bestc,wostc,bestm,(meanc-bestm))
}

