package main

import (
    "github.com/anz-bank/sysl/pkg/sysl"
    "github.com/anz-bank/sysl/pkg/parse"
    "github.com/spf13/afero"
    "os"
    "fmt"
)

var fs = afero.NewOsFs()

func main() {
    args := os.Args[1:]
    aPath := args[0]
    bPath := args[1]
    
    parser := parse.NewParser()
    a, errA := parser.ParseFromFs(aPath, fs)
    b, errB := parser.ParseFromFs(bPath, fs)
    if errA != nil {
        fmt.Printf("Failed to parse file %s", aPath)
    }
    if errB != nil {
        fmt.Printf("Failed to parse file %s", bPath)
    }

    for appName, app := range a.GetApps() {
        bapps := b.GetApps()
        if otherApp, ok := bapps[appName]; ok {
            SyslAppDiff(app, otherApp, appName)
        } else {
            fmt.Printf("> App %s not found\n", appName)
        }
    }
    for appName, _:= range b.GetApps() {
        aapps := a.GetApps()
        if _, ok := aapps[appName]; !ok {
            fmt.Printf("< App %s not found\n", appName)
        }
    }
}

func SyslAppDiff(a *sysl.Application, b *sysl.Application, appName string) {
    atypes := a.GetTypes()
    btypes := b.GetTypes()
    for name, ta := range atypes {
        if tb, ok := btypes[name]; ok {
            diffs := SyslTypeDiffs(ta, tb)
            if len(diffs) > 0 {
                fmt.Printf("In Type: %s.%s\n", appName, name)
                for _,d := range diffs {
                    fmt.Printf("    %s\n", d)
                }
            }
        } else {
            fmt.Printf("> Type %s missing", name)
        }
    }
    for name, _ := range btypes {
        if _,ok := atypes[name]; !ok {
            fmt.Printf("< Type %s missing", name)
        }
    }
}

func SyslTypeDiffs(a *sysl.Type, b *sysl.Type) []string {
    res := make([]string, 0)
    if pa := a.GetPrimitive(); pa != sysl.Type_NO_Primitive {
        if pb := b.GetPrimitive(); pb != sysl.Type_NO_Primitive {
            if pa != pb {
                res = append(res, fmt.Sprintf("Type is [%v] vs [%v]", pa, pb))
            }
        } else {
            res = append(res, "Sysl type mismatch: Primitive expected")
        }
    } else if ra := a.GetTypeRef(); ra != nil {
        if rb := b.GetTypeRef(); rb != nil {
            for i,p := range ra.Context.Path {
                if i >= len(rb.Context.Path) || rb.Context.Path[i] != p {
                    res = append(res, fmt.Sprintf("Reference types do not match: [%v] [%v]",
                        ra.Context.Path, rb.Context.Path))
                    break
                }
            }
        } else {
            res = append(res, "Sysl type mismatch: Reference expected")
        }
    } else if ta := a.GetTuple(); ta != nil {
        if tb := b.GetTuple(); tb != nil {
            for n, fa := range ta.AttrDefs {
                if fb, ok := tb.AttrDefs[n]; ok {
                    diffs := SyslTypeDiffs(fa, fb)
                    if len(diffs) > 0 {
                        res = append(res, fmt.Sprintf("Fields don't match: %s", n))
                        for _,d := range diffs {
                            res = append(res, "  " + d)
                        }
                    }
                } else {
                    res = append(res, fmt.Sprintf("> Field %s missing", n))
                }
            }
            for n, _ := range tb.AttrDefs {
                if _, ok := ta.AttrDefs[n]; !ok {
                    res = append(res, fmt.Sprintf("< Field %s missing", n))
                }
            }
        } else {
            res = append(res, "Sysl type mismatch: Tuple expected")
        }
    } else if sa := a.GetSequence(); sa != nil {
        if sb := b.GetSequence(); sb != nil {
            return SyslTypeDiffs(sa, sb)
        } else {
            res = append(res, "Sysl type mismatch: Sequence expected")
        }
    } else {
        res = append(res, "Unimplemented test for SYSL Type")
    }
    return res
}