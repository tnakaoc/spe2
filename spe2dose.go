package main
import "fmt"
import "strings"
import "strconv"
import "math"
import "os"
import "os/exec"
func main() {
	if len(os.Args)<=2 {
		fmt.Println("usage : ",os.Args[0]," nf=[neutron spectrum] gf=[gamma spectrum]")
		return
	}
	var nfn string
	var gfn string
	for _,v:=range(os.Args) {
		switch v[0:3] {
			case "nf=":
				nfn=v[3:]
			case "gf=":
				gfn=v[3:]
		}
	}
	check:=func(fn string)error{
		fs,err:=os.Open(fn)
		if err==nil {
			fs.Close()
		}
		return err
	}
	err:=check(nfn)
	if err!=nil {
		fmt.Println(err)
		return
	}
	err=check(gfn)
	if err!=nil {
		fmt.Println(err)
		return
	}
	evaluate:=func(ev string,fn string) string{
		res,err:=exec.Command(ev,fn).Output()
		if err!=nil {
			fmt.Println(err)
			return ""
		}
		return string(res)
	}
	nres:=evaluate("nspe2dose",nfn)
	nval:=strings.Split(nres,"\n")
	if !strings.Contains(nval[1],"(1)") {
		fmt.Println(nfn," may not be a neutron spectrum file.")
		return
	}
	nsum,nerr:=func()(float64,float64){
		sta,sto:=strings.Index(nval[1],":"),strings.Index(nval[1],"(+")
		if sta<0||sto<0 {
			return 0.0,0.0
		}
		buf:=strings.Fields(nval[1][sta+1:sto-1])
		return func()(float64,float64){
			cnv:=func(s string)float64{
				ret,er:=strconv.ParseFloat(s,64)
				if er==nil { return ret }
				return math.NaN()
			}
			return cnv(buf[0]),cnv(buf[2])
		}()
	}()
	fmt.Println(nsum,nerr)
	gres:=evaluate("gspe2dose",gfn)
	fmt.Print(gres)
	return
}
