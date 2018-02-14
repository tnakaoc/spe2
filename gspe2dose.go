package main
import "fmt"
import "os"
import "strings"
import "bufio"
import "math"
func main(){
	if len(os.Args)==1 {
		fmt.Println("usage : ",os.Args[0]," [filename]")
		return
	}
	kerm,err := func()([]float64,error){
		file,er := os.Open("/home/nakao/local/gkerma.dat")
		if er != nil {
			return []float64{},er
		}
		defer file.Close()
		scanner:=bufio.NewScanner(file)
		var buf []float64
		for scanner.Scan() {
			var t float64
			_,er:=fmt.Sscanf(scanner.Text(),"%f %f",&t,&t)
			if er!=nil {
				return buf,nil
			}
			buf=append(buf,t)
		}
		return buf,nil
	}()
	if err!=nil {
		fmt.Println("cannot open \"kerma.dat\"")
		return
	}
	file,err:=os.Open(os.Args[1])
	if err!=nil {
		fmt.Println("cannot open ",os.Args[1])
		return
	}
	defer file.Close()
	value:=func()[][4]float64{
		var res [][4]float64
		scanner:=bufio.NewScanner(file)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(),"e-lower"){
				break
			}
		}
		for scanner.Scan() {
			if len(scanner.Text())==0||(scanner.Text())[0]=='#' {
				break
			}
			var buf [4]float64
			_,e:=fmt.Sscanf(scanner.Text(),"%f %f %f %f",&buf[0],&buf[1],&buf[2],&buf[3])
			if e!=nil {
				break
			}
			res=append(res,buf)
		}
		return res
	}()
	dose := 0.0
	derr := 0.0
	for _,buf:=range(value) {
		esta := int(buf[0]*1000)
		esto := func()int {
			e:=int(buf[1]*1000)
			if e<int(len(kerm)) { return e }
			return int(len(kerm))
		}()
		if esta==0 { continue }
		R:=1./(float64(esto-esta))
		for i:=esta;i<esto;i++ {
			dose+=kerm[i-1]*buf[2]*R
			derr+=(kerm[i-1]*buf[2]*buf[3])*(kerm[i-1]*buf[2]*buf[3])*R
		}
	}
	fmt.Println(dose*3600*1e-12,math.Sqrt(derr)*3600*1e-12)
	return
}
