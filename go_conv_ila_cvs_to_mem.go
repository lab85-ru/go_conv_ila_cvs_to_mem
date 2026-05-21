package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"math"
	"runtime/debug"
)

func main() {

	args_n := len(os.Args)
	if args_n == 1 {
		fmt.Println("--------------------------------------------------------------------------------")
		print_git_commit_info()
		fmt.Println(" go_conv_ila_cvs_to_mem <in CSV file> <out MEM file>")
		fmt.Println("--------------------------------------------------------------------------------")
		os.Exit(0)
	}

	if args_n < 3 {
		fmt.Println("ERROR: input parameter.")
		os.Exit(0)
	}
	/*
		fmt.Println("Input parameters.")
		for i := 0; i < args_n; i++ {
			fmt.Println( os.Args[ i ] )
		}
	*/
	file_name_in_str := os.Args[1]
	file_name_out_str := os.Args[2]

	file_in, err := os.Open(file_name_in_str)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file_in.Close()

	reader := csv.NewReader(file_in)

	// Весь файл переносим в массив-массива строк(распарсинный cvs) для последующей обработки
	var array_str [][]string

	// массив максимальных значений для последующего определения разрядности выходных данных
	record_numer_max := make([]int, 0, 100) // Длина 0, вместимость 100(максимальное количество записей в исходном ILA файле)
	record_len := 0
	flag_record_len := 0

	// Read/Parse CVS file ----------------------------------------------------------------------
	for {
		record, e := reader.Read()
		if e != nil {
			fmt.Println(e)
			break
		}

		// Добавляем каждую распарсенную строку в массив
		array_str = append(array_str, record)

		//fmt.Println(record)
		//fmt.Printf("type: %T\n", record)

		if flag_record_len == 0 {
			flag_record_len = 1
			record_len = len(record)
			// Задаем размер массива по количеству записей из файла
			record_numer_max = make([]int, record_len)
		}

		for i := 0; i < record_len; i++ {
			n, err1 := strconv.Atoi(record[i])

			// Самая первая запись это названия полей - пропускаем
			// Если не числа то пропускаем
			if err1 != nil {
				fmt.Println("SKIP value not numer: ", record[i])
				continue
			}

			// Поиск максимального числа
			if n > record_numer_max[i] {
				record_numer_max[i] = n
			}
		}

	}

	fmt.Println("Максимальные значения : ", record_numer_max)
	//fmt.Println(array_str)

	// массив: необходимое количество бит для каждой записи
	var record_numer_bits = make([]int, record_len)

	// Расчитываем количество бит необходимое для размешения данных из полей
	for i := 0; i < record_len; i++ {
		record_numer_bits[i] = calc_veriable_value_len_in_bit(record_numer_max[i])
		//fmt.Println("->", x, " ", y)
	}
	fmt.Println("Требуемая разрядность : ", record_numer_bits)


	// Конечная разрядность памяти verilog
	output_bits := 0;

	for i := 0; i < record_len; i++ {
		output_bits = output_bits + record_numer_bits[i];
	}

	//----------------------------------------------------------------------
	// Создание выходного файла
	file_out, err := os.Create(file_name_out_str) // For read access.
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file_out.Close()

	file_out.WriteString("@00000000")

	for i:=0; i<len(array_str); i++ {
		s := ""

		//fmt.Println()

		for j:=0; j<record_len; j++ {

			//fmt.Println("array_str[i][j]", array_str[i][j])

			val_i, err := strconv.Atoi( array_str[i][j] )

	   		if err != nil {
	   			//fmt.Println("SKIP value not numer")
	   			continue
	   		}

			//fmt.Println(">", uint(val_i), uint(record_numer_bits[j]))
			s = s + conv_int_to_str_bits(uint(val_i), uint(record_numer_bits[j]))
			if j < record_len - 1 {
				s = s + "_";
			}
		}
		file_out.WriteString(s + "\r\n")
	}

	//----------------------------------------------------------------------
	// Создание verilog файла для вставки в Test Bench
	file_out_v, err := os.Create(file_name_out_str+".vh")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file_out_v.Close()

	memory_adr_bits := calc_veriable_value_len_in_bit(len(array_str))

	var str string

	str = fmt.Sprintf("// This file generation from: go_conv_ila_cvs_to_mem\r\n")
	file_out_v.WriteString(str)

	str = fmt.Sprintf("// Vivado ILA(CSV) -> Verilog/ModelSim memory file\r\n")
	file_out_v.WriteString(str)

	str = fmt.Sprintf("\r\n")
	file_out_v.WriteString(str)

	str = fmt.Sprintf("reg [%d:0] ram_adr = 0;\r\n", memory_adr_bits-1)
	file_out_v.WriteString(str)

	str = fmt.Sprintf("reg [%d:0] ram_mem [ %d : 0];\r\n", output_bits-1, len(array_str)-1)
	file_out_v.WriteString(str)

	str = fmt.Sprintf("reg [%d:0] ram_data = 0;\r\n", output_bits-1)
	file_out_v.WriteString(str)
	
	str = fmt.Sprintf("localparam RAM_DATA_END = %d;\r\n", len(array_str)-1)
	file_out_v.WriteString(str)

	str = fmt.Sprintf("\r\n")
	file_out_v.WriteString(str)
	
	str = fmt.Sprintf("initial\r\n")
	file_out_v.WriteString(str)

	str = fmt.Sprintf("begin\r\n")
	file_out_v.WriteString(str)

	str = fmt.Sprintf("    $display(\"load RAM from file: %s\");\r\n", file_name_out_str)
	file_out_v.WriteString(str)

	str = fmt.Sprintf("    $readmemb(\"%s\", ram_mem);\r\n", file_name_out_str)
	file_out_v.WriteString(str)

	str = fmt.Sprintf("end\r\n")
	file_out_v.WriteString(str)

}

//------------------------------------------------------------------------------
// Расчитываем количество бит необходимое для размешения данных из полей
//------------------------------------------------------------------------------
func calc_veriable_value_len_in_bit (n int) int {
	x := math.Log2(float64(n))
	y := int(x + 1)
	return y
}

//------------------------------------------------------------------------------
// Функция конвертации числа с заданной разрядностью в строку битов
//------------------------------------------------------------------------------
func conv_int_to_str_bits(n uint, n_bit uint) string {

	var s string = ""

	var i uint

	for i = 1 << (n_bit - 1); i != 0; i = i >> 1 {
		//fmt.Println("n = ", n, "i = ", i)
		if n & i == 0 {
			s = s + "0"
		} else {
			s = s + "1"
		}
	}

    return s
}

//------------------------------------------------------------------------------
// Функция вывода информации о текущем комите
//------------------------------------------------------------------------------
func print_git_commit_info() {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				fmt.Printf("Git Commit: %s\n", setting.Value)
			}
		}
	}
}