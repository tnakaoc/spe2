package main
import "fmt"
import "os"
import "strings"
import "bufio"
import "math"
func main(){
	if len(os.Args)==1 {
		fmt.Println("usage : ",os.Args[0]," [filename] ([opt])")
		return
	}
	fid,opt:=func()(int,int) {
		id:=0
		op:=0
		for i,s:=range(os.Args){
			if s[0]=='-' {
				switch(s){
					case "-fnr":
						op|=1<<0
					case "-dose":
						op|=1<<1
				}
			} else {
				id=i
			}
		}
		return id,op
	}()
	if fid==0 {
		fmt.Println("usage : ",os.Args[0]," [filename] ([opt])")
		return
	}
	if opt!=0 {
		fmt.Println("option is not yet implemented, sorry.")
	}
	path:="/home/nakao/local/"
	load_file:=func(fnm string)([]float64,error){
		file,er := os.Open(path+fnm)
		if er != nil {
			return []float64{},er
		}
		defer file.Close()
		scanner:=bufio.NewScanner(file)
		var buf []float64
		for scanner.Scan() {
			if scanner.Text()[0]=='#' {
				continue
			}
			var t float64
			_,er:=fmt.Sscanf(scanner.Text(),"%f %f",&t,&t)
			if er!=nil {
				if len(buf)==0 {
					return buf,fmt.Errorf("invalid file format: \"%s\"",fnm)
				}
				return buf,nil
			}
			buf=append(buf,t)
		}
		return buf,nil
	}
	kerm,err := load_file("nkerma.dat")
	if err!=nil {
		fmt.Println(err)
		return
	}
	rbe,err := load_file("nrbe.dat")
	if err!=nil {
		fmt.Println(err)
		return
	}
	file,err:=os.Open(os.Args[fid])
	if err!=nil {
		fmt.Println("cannot open ",os.Args[1])
		return
	}
	defer file.Close()
	e2index:=func(e float64)int{
		if e<0.5e-6 { return 0 }
		if e<10e-3  { return 1 }
		return 2
	}
	sum :=[3]float64{0.0,0.0,0.0}
	serr:=[3]float64{0.0,0.0,0.0}
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
			index:=e2index(buf[0])
			sum[index]+=buf[2]
			serr[index]+=math.Pow(buf[2]*buf[3],2.0)
		}
		return res
	}()
	var dose [2][3]float64
	var derr [2][3]float64
	e2id:=func(e float64)int {
		return int((math.Log10(e*1e11))*1000)
	}
	for _,buf:=range(value) {
		esta := e2id(buf[0])
		esto := func()int {
			e:=e2id(buf[1])
			if e<int(len(kerm)) { return e }
			return int(len(kerm))
		}()
		if esta==0 { continue }
		R:=1./(float64(esto-esta))
		for i:=esta;i<esto;i++ {
			index:=e2index(buf[0])
			dose[0][index]+=buf[2]*kerm[i-1]*R
			dose[1][index]+=buf[2]*kerm[i-1]*rbe[i-1]*R
			derr[0][index]+=math.Pow(buf[2]*buf[3]*kerm[i-1],2.0)*R
			derr[1][index]+=math.Pow(buf[2]*buf[3]*kerm[i-1]*rbe[i-1],2.0)*R
		}
	}
	integral:=func(arr [3]float64)float64{
		ret:=0.0
		for _,v:=range(arr) {
			ret+=v
		}
		return ret
	}
	Nsum:=integral(sum)
	Nerr:=math.Sqrt(integral(serr))
	for i:=0;i<3;i++ {
		serr[i]=math.Sqrt(serr[i])
		derr[0][i]=math.Sqrt(derr[0][i])
		derr[1][i]=math.Sqrt(derr[1][i])
	}
	errat:=func(a float64,b float64)float64 {
		if b>a { return math.NaN() }
		return b/a*100.0
	}
	bar:="**********"
	bar=bar+bar+bar+"****"
	pdose:=func(lab string,tit [4]string,v [3]float64,ve [3]float64,r float64){
		fmt.Printf("%s\n",lab)
		lsum :=v[0]+v[1]+v[2]
		if v[0]<0.0||v[1]<0.0||v[2]<0.0||lsum==0.0{
			fmt.Println("")
			fmt.Println("> unrecognized error detected.")
			fmt.Println("> maybe, input file is not a spectrum data.")
			return
		}
		lesum:=math.Sqrt(ve[0]*ve[0]+ve[1]*ve[1]+ve[2]*ve[2])
		label:=[4]string{"total   ","thermal ","epi     ","fast    "}
		genval:=func(a0 float64,a1 float64)(float64,float64,float64){
			a,b,c:=a0,a1,errat(a0,a1)
			if b>a {
				b=math.NaN()
			}
			return a,b,c
		}
		a,b,c:=genval(r*lsum,r*lesum)
		fmt.Printf("\t%s%s: %.3e +- %.3e (+-%5.2f %%) | Ratio\n",label[0],tit[0],a,b,c)
		for i:=0;i<3;i++ {
			a,b,c=genval(r*v[i],r*ve[i])
			fmt.Printf("\t%s%s: %.3e +- %.3e (+-%5.2f %%) | (%5.2f %%) ||%s\n",label[i+1],tit[i+1],a,b,c,errat(lsum,v[i]),bar[:int(v[i]/lsum*100.0/3.0)])
		}
	}
	pdose("Neutron [/cm^2/s]"       ,[4]string{" (1)"," (2)"," (3)"," (4)"},sum,serr,1.0)
	pdose("Neutron [/cm^2/h]"       ,[4]string{" (A)"," (B)"," (C)"," (D)"},sum,serr,3600.0)
	pdose("Dose (Neutron) [Gy/h]"   ,[4]string{"(a0)","(b0)","(c0)","(d0)"},dose[0],derr[0],3600.0*1e-12)
	pdose("Dose (Neutron) [Gy-eq/h]",[4]string{"(a0)","(b0)","(c0)","(d0)"},dose[1],derr[1],3600.0*1e-12)
	fmt.Printf("Dose Ratio (Neutron) [Gy cm^2/Neutron]\n")
	pratio:=func(tit [4]string,g float64,dg float64,f float64,df float64,ul bool){
		a,b:=g/f,math.Sqrt((g*g*df*df+f*f*dg*dg)/f/f/f/f)
		r:=errat(a,b)
		if b>a {
			b=math.NaN()
		}
		if ul { fmt.Printf("\x1b[4m") }
		fmt.Printf("\t%-8s/ %-5s %4s/%4s: %.3e +- %.3e (+-%5.2f %%)",tit[0],tit[1],tit[2],tit[3],a,b,r)
		if ul { fmt.Printf("\x1b[0m") }
		fmt.Printf("\n")
	}
	pratio([4]string{"thermal","epi"  ,"(b0)","(C)"},dose[0][0]*1e-12,derr[0][0]*1e-12,sum[1],serr[1],false)
	pratio([4]string{"epi"    ,"epi"  ,"(c0)","(C)"},dose[0][1]*1e-12,derr[0][1]*1e-12,sum[1],serr[1],false)
	pratio([4]string{"fast"   ,"epi"  ,"(d0)","(C)"},dose[0][2]*1e-12,derr[0][2]*1e-12,sum[1],serr[1],true )
	pratio([4]string{"thermal","total","(b0)","(A)"},dose[0][0]*1e-12,derr[0][0]*1e-12,Nsum  ,Nerr   ,false)
	pratio([4]string{"epi"    ,"total","(b0)","(A)"},dose[0][1]*1e-12,derr[0][1]*1e-12,Nsum  ,Nerr   ,false)
	pratio([4]string{"fast"   ,"total","(b0)","(A)"},dose[0][2]*1e-12,derr[0][2]*1e-12,Nsum  ,Nerr   ,false)
	return
}
