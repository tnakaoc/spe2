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
	file,err:=os.Open(os.Args[1])
	if err!=nil {
		fmt.Println("cannot open ",os.Args[1])
		return
	}
	defer file.Close()
	loadfile:=func(fnm string)([][4]float64,error){
		fil,err:=os.Open(fnm)
		var res [][4]float64
		if err!=nil {
			return res,err
		}
		defer fil.Close()
		scanner:=bufio.NewScanner(fil)
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
			if buf[0]==0.0 { continue }
			if e!=nil {
				break
			}
			res=append(res,buf)
		}
		return res,nil
	}
	path:="/home/nakao/bsas/spe2zdis_data"
	nG,nS,bS,gS,err:=func()([][]float64,[][]float64,[][]float64,[][]float64,error){
		var rnG [][]float64
		var rnS [][]float64
		var rbS [][]float64
		var rgS [][]float64
		for i:=0;i<200;i++ {
			fnm:=fmt.Sprintf("%s/%.1f.j",path,float64(i)/10)
			file,err := os.Open(fnm)
			if err!=nil {
				return [][]float64{}, [][]float64{}, [][]float64{}, [][]float64{}, err
			}
			defer file.Close()
			scan:=bufio.NewScanner(file)
			var tnG []float64
			var tnS []float64
			var tbS []float64
			var tgS []float64
			for scan.Scan() {
				var buf [5]float64
				n,_:=fmt.Sscan(scan.Text(),&buf[0],&buf[1],&buf[2],&buf[3],&buf[4])
				if n==5 {
					tnG=append(tnG,buf[1])
					tnS=append(tnS,buf[2])
					tbS=append(tbS,buf[3])
					tgS=append(tgS,buf[4])
				}
			}
			rnG=append(rnG,tnG)
			rnS=append(rnS,tnS)
			rbS=append(rbS,tbS)
			rgS=append(rgS,tgS)
		}
		return rnG,rnS,rbS,rgS,nil
	}()
	if err!=nil {
		fmt.Println("cannot load spectrum database")
		fmt.Println(err)
		return
	}
	fgen:=func(e float64,z float64,db [][]float64)float64{
		iz:=uint(z*10)
		if iz>200||e>100.0 { return 0.0 }
		id:=func(ev float64)int {
			//return int((math.Log10(ev*1e9))*100)
			return int((math.Log10(ev*1e12))*30.0)
		}(e)
		id2e:=func(id int)float64 {
			//return math.Pow(10.0,float64(id)/100.0)*1e-9
			return math.Pow(10.0,float64(id)/30.0)*1e-12
		}
		if id<0 { id = 0 }
		x0:=math.Log(float64(id2e(id  )))
		x1:=math.Log(float64(id2e(id+1)))
		y0:=math.Log(db[iz][id  ])
		y1:=math.Log(db[iz][id+1])
		le:=math.Log(e)
		return math.Exp((le-x0)*(y0-y1)/(x0-x1)+y0)
	}
	nGy:=func(e float64,z float64)float64{ return fgen(e,z,nG) }
	nSv:=func(e float64,z float64)float64{ return fgen(e,z,nS) }
	bSv:=func(e float64,z float64)float64{ return fgen(e,z,bS) }
	gSv:=func(e float64,z float64)float64{ return fgen(e,z,gS) }
	output:=func(fnm string,ofn string){
		value,err:=loadfile(fnm)
		if err!=nil {
			fmt.Println(err)
			return
		}
		if len(value)==0 {
			fmt.Println("unrecognized error")
			fmt.Println("maybe, input file is not a spectrum data")
			return
		}
		ofs,_:=os.Create(ofn)
		is_lethargy:=false
		{
			fmt.Fprintf(ofs,"# \n")
			fmt.Fprintf(ofs,"# Created by \"%s\", from spectrum file \"%s\"\n",os.Args[0],fnm)
			fmt.Fprintf(ofs,"# \n")
			fmt.Fprintf(ofs,"# Input file information.\n")
			fmt.Fprintf(ofs,"# [[%s]]\n",fnm)
			fmt.Fprintf(ofs,"#\t%-10s\t%-10s\t%-10s\t%-10s\n","e-lower","e-upper","neutron","err")
			sum:=0.0
			pret:=0.0
			for _,v:=range(value) {
				fmt.Fprintf(ofs,"#\t%-10.4E\t%-10.4E\t%-10.4E\t%-10.4E\n",v[0],v[1],v[2],v[3])
				sum+=v[2]
				if !is_lethargy {
					rat:=v[1]/v[0]
					if pret!=0.0&&math.Abs(pret/rat-1.0)>0.01 {
						is_lethargy=true
					}
					pret=v[1]/v[0]
				}
			}
			fmt.Fprintf(ofs,"# \n")
			fmt.Fprintf(ofs,"#\t%-10s\t%-10s\t%-10.4E\n","sum over"," ",sum)
			fmt.Fprintf(ofs,"# \n")
		}
		fmt.Fprintf(ofs,"# Z-distribution information.\n")
		fmt.Fprintf(ofs,"#\t%-12s\t%-12s\t%-12s\t%-12s\t%-12s\n","Z[cm]","N [Gy/h]","N [Gy-eq/h]","B [Gy/h/ppm]","G [Gy/h]")
		for i:=0;i<200;i++ {
			var val [4]float64
			z:=float64(i)/10
			for _,v:=range(value) {
				evalue  :=math.Exp((math.Log(v[0])+math.Log(v[1]))*0.5)
				lethargy:=func()float64{
					if is_lethargy { return math.Log(v[1]/v[0]) }
					return 1.0
				}()
				val[0]+=v[2]*(nGy(evalue,z))*lethargy
				val[1]+=v[2]*(nSv(evalue,z))*lethargy
				val[2]+=v[2]*(bSv(evalue,z))*lethargy
				val[3]+=v[2]*(gSv(evalue,z))*lethargy
			}
			fmt.Fprintf(ofs,"\t%-12.5E\t%-12.5E\t%-12.5E\t%-12.5E\t%-12.5E\n",z,val[0],val[1],val[2],val[3])
		}
		fmt.Fprintf(ofs,"#\n")
		fmt.Fprintf(ofs,"# [EOF]\n")
	}
	scanner:=bufio.NewScanner(file)
	scanner.Scan()
	if scanner.Text()=="[[filelist]]" {
		for scanner.Scan() {
			line:=func()string{
				if strings.Contains(scanner.Text(),"#") {
					return scanner.Text()[:strings.Index(scanner.Text(),"#")]
				}
				return scanner.Text()
			}()
			field:=strings.Fields(line)
			if len(field)==2 {
				output(field[0],field[1])
			}
		}
	} else {
		output(os.Args[1],func()string{
				if len(os.Args)==3 {
					return os.Args[2]
				}
				return "/dev/stdout"
			}())
	}
	return
}
