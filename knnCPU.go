package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
)

//ReadCsv es la lectura del dataset en csv
func ReadCsv(file string, v *[]Vec) {
	csvfile, err := os.Open(file)
	if err != nil {
		log.Fatalln("No se encuentra el archivo.")
	}

	r := csv.NewReader(csvfile)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		newVec := Vec{}        //Fila vacia
		numVals := []float64{} //Arrelgo de valores de la fila (creado vacio)

		for i, val := range record {
			if i < len(record)-1 {
				n, _ := strconv.ParseFloat(val, 64)
				numVals = append(numVals, n)
			} else {
				newVec.class = val //Determinamos la clase de la fila
			}
		}
		newVec.elements = numVals  //Determinamos los elementos de la fila
		newVec.size = len(numVals) //Determinamos la cantidad de elementos de la fila
		*v = append(*v, newVec)
	}
}

//DistanceCalculation es el calculo de distancia que da aescoger entre Euclidiana o Manhattan
func (v *Vec) DistanceCalculation(toV Vec, dist string) float64 {
	sums := []float64{}
	var sum float64
	var result float64

	for i := 0; i < v.size; i++ {
		sums = append(sums, v.elements[i]-toV.elements[i])
	}
	//Validacion de si es M/m de Manhattan
	if dist == "M" || dist == "m" {
		for i := 0; i < len(sums); i++ {
			sum += sums[i]
		}
		result = math.Abs(sum)
		//Validacion de si es E/e de Euclidiana
	} else if dist == "E" || dist == "e" {
		for i := 0; i < len(sums); i++ {
			sum += math.Pow(sums[i], 2)
		}
		result = math.Sqrt(sum)
	} else {
		log.Fatalln("Indicador de distancia equivocado.")
	}
	return result
}

//SubKnn retorna las K instancias mas cercanas
func SubKnn(k int, newInstance Vec, datasetInstances []Vec) []Vec {
	if k%2 == 0 {
		log.Fatalln("K no puede ser par.")
	}

	almacenDistance := []float64{}
	var distance float64

	distanceIndicator := "e" //Using Manhattan (for now)
	m := make(map[float64]Vec)

	for i := 0; i < len(datasetInstances); i++ {
		distance = newInstance.DistanceCalculation(datasetInstances[i], distanceIndicator)
		almacenDistance = append(almacenDistance, distance)
		m[distance] = datasetInstances[i] //Key -> Distance // Value -> Dataset Instance
	}

	sort.Float64s(almacenDistance) //Menor a mayor sort
	var aux int

	if len(datasetInstances) < k {
		aux = len(datasetInstances)
	} else {
		aux = k
	}
	kNearestSlice := make([]Vec, aux)
	for i := 0; i < aux; i++ {
		kNearestSlice[i] = m[almacenDistance[i]]
	}
	return kNearestSlice
}

//Knn utiliza las k instancias retornadas por subKNN para obtener la clase predecida
func Knn(k int, kNearestSlice []Vec) string {

	if k%2 == 0 {
		log.Fatalln("K no puede ser par.")
	}

	var predClass string

	classCount := make(map[string]int)

	for i := 0; i < k; i++ {
		classCount[kNearestSlice[i].class]++
	}

	aux := 0

	for strClass, intCount := range classCount {
		if aux < intCount {
			aux = intCount
			predClass = strClass
		}
	}

	return predClass
}

//Vec es una fila del dataset (tamano, elementos numericos y clase)
type Vec struct {
	size     int
	elements []float64
	class    string
}

//strVec es lo mismo que Vec solo que los elementos estan en formato string
type strVec struct {
	size           int
	stringElements []string
	class          string
}

//DivisorDeVecs divide el dataset
func DivisorDeVecs(vecs []Vec, nodos int) [][]Vec {
	cantidadInstancias := len(vecs)
	//almacenVecs := make([][]Vec, nodos)
	var almacenVecs [][]Vec
	tamano := (cantidadInstancias + nodos - 1) / nodos

	for i := 0; i < cantidadInstancias; i += tamano {
		end := i + tamano
		if end > cantidadInstancias {
			end = cantidadInstancias
		}
		almacenVecs = append(almacenVecs, vecs[i:end])
	}
	return almacenVecs
}

//vecToStrVec convierte los elementos a string
func vecToStrVec(vecs []Vec) []strVec {
	newStrVec := make([]strVec, len(vecs))
	var auxStrAlm []string
	for i := 0; i < len(vecs); i++ {
		newStrVec[i].size = vecs[i].size
		newStrVec[i].class = vecs[i].class

		for j := 0; j < len(vecs[i].elements); j++ {

			var auxStr string
			auxStr = strconv.FormatFloat(vecs[i].elements[j], 'f', 6, 64)
			auxStrAlm = append(auxStrAlm, auxStr)
		}
		newStrVec[i].stringElements = auxStrAlm
		auxStrAlm = nil
	}
	return newStrVec
}

//strVecToVec convierte los elementos a float64
func strVecToVec(strVecs []strVec) []Vec {
	newVec := make([]Vec, len(strVecs))
	var auxFltAlm []float64
	for i := 0; i < len(strVecs); i++ {
		newVec[i].size = strVecs[i].size
		newVec[i].class = strVecs[i].class

		for j := 0; j < len(strVecs[i].stringElements); j++ {

			var auxFlt float64
			auxFlt, _ = strconv.ParseFloat(strVecs[i].stringElements[j], 64)
			auxFltAlm = append(auxFltAlm, auxFlt)
		}
		newVec[i].elements = auxFltAlm
		auxFltAlm = nil
	}
	return newVec
}

func main() {

	vecs := []Vec{}
	ReadCsv("Prostate_Cancer_No_NA.csv", &vecs)

	var n int
	fmt.Print("Ingrese la cantidad de nodos: ")
	fmt.Scanf("%d\n", &n)

	if n > len(vecs) {
		log.Fatalln("La cantidad de nodos no puede ser mayor a la cantidad de instancias.")
	}

	var k int
	fmt.Print("Ingrese el valor de K: ")
	fmt.Scanf("%d\n", &k)

	almacenVecsGrupos := DivisorDeVecs(vecs, n)

	for i := 0; i < len(almacenVecsGrupos); i++ {
		/*fmt.Println("Datos del nodo", i+1)
		fmt.Println(almacenVecsGrupos[i])
		fmt.Println(len(almacenVecsGrupos[i]))
		fmt.Println("")*/
	}

	storageElements := []float64{23, 12, 151, 954, 0.143, 0.278, 0.18, 0.07}

	newInstance := Vec{vecs[0].size, storageElements, "NC"} //NC es NO CLASS (Al ser nueva, no tiene clase)

	//Concurrencia
	var bestSubKnn [][]Vec
	channel := make(chan []Vec)
	for i := 0; i < n; i++ {
		go func(differentI int) {
			aux := SubKnn(k, newInstance, almacenVecsGrupos[differentI])
			/*fmt.Println("Instancias mas cercanas del nodo", differentI+1)
			fmt.Println(aux)
			fmt.Println(len(aux))
			fmt.Println("")*/
			channel <- aux
		}(i)
		bestSubKnn = append(bestSubKnn, <-channel)
	}
	//Fin de concurrencia
	var bestKNN []Vec

	for i := 0; i < len(bestSubKnn); i++ {
		bestKNN = append(bestKNN, bestSubKnn[i]...)
	}
	predClass := Knn(k, bestKNN)

	testStr := vecToStrVec(vecs)
	testFlt := strVecToVec(testStr)

	fmt.Println("Clase predecida: ", predClass)
	fmt.Println(testStr)
	fmt.Println(testFlt)
}
