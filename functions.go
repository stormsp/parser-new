package main

import (
	PKG "AlgorithmsRabbit/connections"
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// TODO: nmin, nmax,dost, true,false, cyclesec,execsec,
type RepsValue interface{}

type Rep1 struct {
	sys_num int
	SYS_NUM int
	Value   bool
}

// Константа, представляющая количество миллисекунд в одном тике
const TickDuration float32 = 55.0

//var Reps map[string]Rep

var database = map[any]bool{
	"a": true,
	"b": false,
}

const (
	// Предположим, что период цикла в тиках таймера задан как константа
	ticksPerCycle uint = 100 // количество тиков в одном цикле

	// Количество тиков в секунде, предположим, что 1 секунда = 100 тиков
	ticksPerSecond uint = 100
)

//работа переводчика____________________________________________________

func findReps(text string) map[string]Rep {
	re := regexp.MustCompile(`\{([^{}\n]+)\}`)
	matches := re.FindAllStringSubmatch(text, -1)

	reps := make(map[string]Rep)

	for _, match := range matches {
		if len(match) > 1 {
			rep := match[1]
			// Убираем пробелы и символы табуляции в начале и конце строки
			rep = regexp.MustCompile(`^\s+|\s+$`).ReplaceAllString(rep, "")

			// Проверяем, был ли репер добавлен ранее
			if _, found := reps[rep]; !found {
				// Генерируем случайное значение 0 или 1
				randomValue := rand.Intn(2)
				if randomValue == 0 {
					reps[rep] = Rep{Value: 0}
				} else {
					reps[rep] = Rep{Value: 1}
				}
			}
		}
	}

	return reps
}

func ReplaceExpressions(text string, reps map[string]Rep) string {
	re := regexp.MustCompile(`\{([^{}\n]+)\}`)

	// Заменяем выражения в тексте
	result := re.ReplaceAllStringFunc(text, func(match string) string {
		repName := match[1 : len(match)-1] // Извлекаем имя репера из скобок
		if _, found := reps[repName]; found {
			return fmt.Sprintf("val(\"%s\")", repName)
		}
		return match // Если репер не найден, оставляем выражение без изменений
	})

	return result
}

func ReplaceAllStringRegexp(input, pattern, replace string) string {
	reg := regexp.MustCompile(pattern)
	return reg.ReplaceAllString(input, replace)
}

func ReplaceAllStringRegexpFunc(input, pattern string, repl func(string) string) string {
	reg := regexp.MustCompile(pattern)
	return reg.ReplaceAllStringFunc(input, repl)
}

//работа переводчика____________________________________________________

// математические функции
// NMIN
func NMIN(values ...float32) (float32, error) {
	if len(values) == 0 {
		return 0, fmt.Errorf("no values provided")
	}
	min := float32(math.Inf(1)) // Инициализируем min как положительную бесконечность
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min, nil
}

// NMAX
func NMAX(values ...float32) (float32, error) {
	if len(values) == 0 {
		return 0, fmt.Errorf("no values provided")
	}
	max := float32(math.Inf(-1)) // Инициализируем max как отрицательную бесконечность
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max, nil
}

//логические функции
// dost проверяет достоверность переменной по её имени

func dost(Raper string) bool {
	PKG.InputMap.Mu.Lock()
	value := PKG.InputMap.Out[Raper].Reliability
	PKG.InputMap.Mu.Unlock()
	return value
}

// TRUE всегда возвращает true, независимо от входных аргументов
func TRUE(args ...interface{}) bool {
	return true
}

// FALSE принимает любое количество аргументов и всегда возвращает false
func FALSE(args ...interface{}) bool {
	return false
}

// битовые функции
// BITS получает значение группы битов по маске, начиная с заданного бита
func BITS(dw uint32, bit0 uint, mask uint32) uint32 {
	return (dw >> bit0) & mask
}

// BXCHG переставляет байты в двойном слове в соответствии с заданной последовательностью
func BXCHG(dw uint32, byteseq string) uint32 {
	var result uint32
	for i, char := range byteseq {
		if char >= '1' && char <= '4' {
			shift := (4 - uint(char-'0')) * 8
			result |= ((dw >> shift) & 0xFF) << (3 - i) * 8
		}
	}
	return result
}

// SETBITS устанавливает значения битов в числе на заданное значение
func SETBITS(dw uint32, cnt uint, shf uint, val uint32) uint32 {
	mask := uint32((1<<cnt - 1) << shf) // Создаём маску для установки битов
	return (dw &^ mask) | ((val << shf) & mask)
}

// функции времени выполнения
// CYCLESEC возвращает период запуска алгоритма в секундах
func CYCLESEC() float32 {
	return float32(ticksPerCycle) / float32(ticksPerSecond)
}

// Функция, которую мы хотим замерить
func someTask() {
	// Имитация некоторой длительной операции
	time.Sleep(2 * time.Second)
}

// EXECSEC измеряет и возвращает время выполнения функции someTask в секундах
func EXECSEC() float32 {
	startTime := time.Now()            // Засекаем время начала выполнения
	someTask()                         // Выполнение функции, время которой необходимо измерить
	duration := time.Since(startTime)  // Вычисляем длительность выполнения
	return float32(duration.Seconds()) // Возвращаем длительность в секундах
}

// функции над таймерами
func TIMERMSEC(t time.Time) int {
	return t.Nanosecond() / 1e6
}

// TIMERSEC возвращает секунды от начала времени, указанного в параметре
func TIMERSEC(t time.Time) int {
	return t.Second()
}

// TIMERMIN возвращает минуты от начала времени, указанного в параметре
func TIMERMIN(t time.Time) int {
	return t.Minute()
}

// TIMERHOUR возвращает часы от начала времени, указанного в параметре
func TIMERHOUR(t time.Time) int {
	return t.Hour()
}

// MAKETIMER рассчитывает время счётчика из пользовательских данных
func MAKETIMER(hour, min, sec, msec int) time.Time {
	return time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), hour, min, sec, msec*1e6, time.Local)
}

// Функция GETTICKS, учитывающая константу TickDuration
func GETTICKS(prevTickCnt float32) float32 {
	// Получаем текущее время в миллисекундах с момента запуска программы
	currentTimeMillis := float32(time.Since(startTime).Milliseconds())
	// Конвертируем текущее время в тики
	currentTickCnt := currentTimeMillis / TickDuration
	if prevTickCnt != 0 {
		// Если предыдущее значение счетчика тиков не равно нулю, возвращаем разницу
		return currentTickCnt - prevTickCnt
	}
	// Если предыдущее значение счетчика тиков равно нулю, возвращаем текущее количество тиков
	return currentTickCnt
}

// Измененная функция TICKSIZE, которая возвращает значение, равное 1/18.18
func TICKSIZE() float32 {
	// Засекаем начальное время
	startTime := time.Now()
	// Засекаем конечное время после прошедшей одной секунды
	time.Sleep(1 * time.Second)
	endTime := time.Now()
	// Рассчитываем разницу в секундах и преобразуем её в int
	duration := float32(endTime.Sub(startTime).Seconds())
	// Выводим разницу в секундах
	fmt.Printf("TICKSIZE: %f seconds\n", duration)
	// Возвращаем значение, соответствующее длительности одного тика в секундах
	return duration / (1000.0 / TickDuration)
}

// функции перезагрузки не понял как реализовать на винде
// stopSoftdog имитирует остановку "программного сторожевого таймера".
// Теперь функция также проверяет, работает ли она на Linux, и только тогда выполняет свои действия.
func STOP_SOFTDOG() {
	// Получаем информацию об операционной системе
	osType := runtime.GOOS

	if osType != "linux" {
		fmt.Println("Функция STOP_SOFTDOG поддерживается только на Linux.")
		return
	}

	fmt.Println("STOP_SOFTDOG: Создание файла coredump.txt и завершение работы программы на Linux.")
	file, err := os.Create("coredump.txt")
	if err != nil {
		fmt.Println("Ошибка при создании файла coredump.txt:", err)
		return
	}
	defer file.Close()

	file.WriteString("Coredump caused by software watchdog timer.\n")

	// Эмулируем завершение работы программы
	os.Exit(1)
}

func RESET(param int) {
	switch runtime.GOOS {
	case "windows":
		if param == -1 {
			fmt.Println("Выполнение команды shutdown для Windows.")
			cmd := exec.Command("shutdown", "/s", "/t", "0") // Немедленное выключение
			if err := cmd.Run(); err != nil {
				fmt.Println("Ошибка при выполнении команды shutdown:", err)
			}
		} else {
			fmt.Println("Мягкая перезагрузка не поддерживается на Windows с параметром, отличным от -1.")
		}
	case "linux":
		fmt.Println("Создание файла coredump и мягкая перезагрузка на Linux.")
		_, err := os.Create("coredump.txt")
		if err != nil {
			fmt.Println("Не удалось создать файл coredump.txt:", err)
			return
		}
		fmt.Println("Файл coredump.txt создан успешно.")
		// Имитация деления на ноль для вызова паники
		fmt.Println("Деление на ноль для искусственной перезагрузки.")
		_ = 1 / (param - param) // Паника: деление на ноль
	default:
		fmt.Println("Операционная система не поддерживается.")
	}
}

var vars []float32

func pidreg(
	boi, S0_Ki, S1_Kp, S2_Kd float32,
	g, fu, y, y_diap, umin, umax, Tf, u_vmax, g_vmax float32,
	reg_mode,
	yf_size,
	yf_type, adapt_type, gf_type, def_type, y_trust float32) float32 {
	// Constants
	const tickRate = 0.01 // Tick rate in seconds

	// Initialize internal variables if not already initialized
	if float32(len(vars)) < boi+27 {
		newVars := make([]float32, int(boi)+27)
		copy(newVars, vars)
		vars = newVars
	}

	// Update tick counter
	vars[int(boi)+0] += tickRate

	// Calculate filtered setpoint
	if gf_type == 1 {
		// Limit rate of change for setpoint
		vars[int(boi)+2] += float32(math.Min(float64(g_vmax)*float64(tickRate), math.Abs(float64(g)-float64(vars[int(boi)+2]))) * math.Copysign(1, float64(g)-float64(vars[int(boi)+2])))
	} else if gf_type == 2 {
		// Exponential smoothing
		vars[int(boi)+2] += float32((float32(g) - vars[int(boi)+2]) * float32(1-math.Exp(float64(float32(-tickRate)/Tf))))
	} else {
		// No filter
		vars[int(boi)+2] = g
	}

	// Calculate filtered feedback signal
	if yf_type == 1 {
		// Median filter (for simplicity, nearest to the average is used)
		vars[int(boi)+3] = (vars[int(boi)+3] + y) / 2
	} else if yf_type == 2 {
		// Simple moving average
		vars[int(boi)+18] = (vars[int(boi)+18]*float32(yf_size-1) + y) / float32(yf_size)
		vars[int(boi)+3] = vars[int(boi)+18]
	} else {
		// No filter
		vars[int(boi)+3] = y
	}

	// Calculate error
	vars[int(boi)+4] = vars[int(boi)+2] - vars[int(boi)+3]

	// Adaptation of coefficients
	if adapt_type == 1 && math.Abs(float64(vars[int(boi)+4]/y_diap)) < 0.01 {
		S0_Ki *= 0.5
		S1_Kp *= 0.1
		S2_Kd *= 0.1
	}

	// Calculate PID terms
	vars[int(boi)+9] = S1_Kp * vars[int(boi)+4]                                  // Proportional term
	vars[int(boi)+10] = vars[int(boi)+12] + S0_Ki*vars[int(boi)+4]*tickRate      // Integral term
	vars[int(boi)+11] = S2_Kd * (vars[int(boi)+4] - vars[int(boi)+5]) / tickRate // Derivative term

	// Update integral component
	vars[int(boi)+12] = vars[int(boi)+10]

	// Calculate control signal
	u := vars[int(boi)+9] + vars[int(boi)+10] + vars[int(boi)+11]

	// Apply output limits
	if u > umax {
		u = umax
	} else if u < umin {
		u = umin
	}

	// Apply rate of change limits
	if u_vmax > 0 {
		u += float32(math.Min(float64(u_vmax*tickRate), math.Abs(float64(u-vars[int(boi)+14]))) * math.Copysign(float64(1), float64(u-vars[int(boi)+14])))
	}

	// Save previous error for next derivative calculation
	vars[int(boi)+5] = vars[int(boi)+4]

	// Update internal state
	vars[int(boi)+14] = u

	// Return control signal
	return u
}

//функции алгоритмов управления
//func SET(parameter string, value float32) {
//	val(parameter) = value
//}

func set_wait(parameter string, value float32, timeout float64) float32 {
	timeStart := time.Now()
	for {
		if val(parameter) == value {
			return 1
		}
		if time.Since(timeStart).Seconds() > timeout {
			return 0
		}
		time.Sleep(250 * time.Millisecond)
	}
}

// функции работы с массивами
func FINDOUT(first int, value int, count int, arr []int) int {
	for i := first; i < first+count; i++ {
		if arr[i] == value {
			return i
		}
	}
	return -1 // Возвращаем -1, если элемент не найден
}

// isBool проверяет, является ли значение логическим (bool)
func isBool(val RepsValue) bool {
	_, ok := val.(bool)
	return ok
}

// isInt проверяет, является ли значение целочисленным (int)
func isInt(val RepsValue) bool {
	_, ok := val.(int)
	return ok
}

func reset(param int) error {
	if param == -1 {
		// Выполнить операцию shutdown для Windows
		cmd := exec.Command("shutdown", "/r", "/t", "0")
		return cmd.Run()
	}

	// В противном случае, вам нужно определить логику для мягкой перезагрузки в Linux,
	// например, использование команды kill для отправки сигнала перезагрузки
	// или других подходящих методов.

	return fmt.Errorf("Unsupported operation")
}

func INITOUTS(firstIndex, value, count int) []int {
	// Создаем массив для хранения выходных переменных
	output := make([]int, count)

	// Инициализируем выходные переменные значениями value
	for i := 0; i < count; i++ {
		output[i] = value
	}

	return output
}

// Оператор BEEP - выдает звуковой сигнал (однократный короткий "бип").
func BEEP() {
	fmt.Println("Beep!")
}

// Оператор SIREN_ON - включает звуковой сигнал "сирена".
func SIREN_ON() {
	fmt.Println("Siren ON")
	// Ваш код для включения сирены
}

// Оператор SIREN_OFF - выключает звуковой сигнал "сирена".
func SIREN_OFF() {
	fmt.Println("Siren OFF")
	// Ваш код для выключения сирены
}

// Оператор EXECUTE - выполняет командный файл по указанному пути и имени.
func EXECUTE(path, filename string) error {
	cmd := exec.Command(path, filename)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// Функция для добавления фигурных скобок к конструкциям if
func addBracesToIfStatements(code string) string {
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(code))

	// Регулярное выражение для поиска строк, начинающихся с if
	reIf := regexp.MustCompile(`^\s*if\s*\(?.*\)?\s*$`)

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		// Проверяем, начинается ли строка с if и нет ли уже фигурной скобки
		if reIf.MatchString(trimmedLine) && !strings.HasSuffix(trimmedLine, "{") {
			line += " {"
		}
		result.WriteString(line + "\n")
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading text:", err)
	}

	return result.String()
}

// Библиотеки __________________________________________________________________________________________________________
// valTrack.evl
func valTrack(val1 float32, timeout float32, id float32) float32 {
	//aout[int(id)] = float32(val1)
	//aout[int(id)] = 0
	if (val1) == (0) {
		aout[int(id)] = 0
		return (0)
	}
	fmt.Println("aout[id]", aout[int(id)])
	// aout[id] время перехода в состояние, отличное от 0
	// для вычисления тайм-аута (в тиках со старта зонда)
	if (aout[int(id)]) == (0) {
		aout[int(id)] = GETTICKS(0)
	}
	fmt.Println("getticks aout[int(id)]", GETTICKS(aout[int(id)]))
	fmt.Println("ticks: ", (GETTICKS(aout[int(id)]) * TICKSIZE()))

	if (GETTICKS(aout[int(id)]) * TICKSIZE()) >= (timeout) {
		fmt.Println("Вошел в тикс")
		return (1)
	}

	return (0)
}

// valTrackGt и valTrackLt возращают 0 если отслеживаемый
// параметр не достоверен, или если не нарушена
// граница, или если со времени нарушения
// не прошло timeout секунд. В противном случае
// функции возвращают 1.
func valTrackGt(val1 string, bound float32, timeout float32, id float32) float32 {
	if (dost(val1)) != false && (val(val1) > bound) {
		return (valTrack(val(val1), timeout, id))
	} else {
		return 0
	}
}

func valTrackLt(val1 string, bound float32, timeout float32, id float32) float32 {
	if (dost(val1)) != false && (val(val1) < bound) {
		return (valTrack(val(val1), timeout, id))
	} else {
		return 0
	}
}

// при достоверности одного из трех каналов, недостоверный канал заменяется
// значением параметров с ЭКМ давление на вых низкое, высокое
func valTrackLt_dost(val1 string, bound string, timeout float32, id float32, p_ekm float32) float32 {
	if dost(val1) {
		if val(val1) < val(bound) {
			return (valTrack(float32(1), timeout, id))
		} else {
			return (valTrack(float32(0), timeout, id))
		}
	} else {
		return (p_ekm)
	}
}

func valTrackGt_dost(val1 string, bound string, timeout float32, id float32, p_ekm float32) float32 {
	if dost(val1) {
		if val(val1) > val(bound) {
			return (valTrack(float32(1), timeout, id))
		} else {
			return (valTrack(float32(0), timeout, id))
		}
	} else {
		return (p_ekm)
	}
}

// yahont.evl
func yahont4(a float32, cod string) {
	if dost(cod) {
		if (val(cod) == (0)) || (val(cod) == (1)) || (val(cod) == (2)) { //неопр || кз || обрыв {
			dout[int(a)] = 0 //--неопр
		} else {
			if val(cod) == (3) { //норма {
				dout[int(a)] = 1 //--норма
			} else {
				if val(cod) == (4) { //вним {
					dout[int(a)] = 3
				} else {
					if val(cod) == (5) { //трев {
						dout[int(a)] = 2 //--пожар
					} else {
						dout[int(a)] = 0 //--неисправн
					}
				}
			}
		}
	} else {
		dout[int(a)] = 0
	}
}

// —осто¤ние концевика
// ¬ паспорте должно быть:
// 0- неопределенность, 1-норма, 2-вскитые
//

func yahont41(a int, cod any) any {
	if (cod) == (0) { // код 0-неопр {
		dout[a] = 0
	} else {
		if (cod) == (3) { // код 3-норма {
			dout[a] = 1
		} else {
			if (cod) == (6) { // код 6-трев {
				dout[a] = 2
			} else {
				dout[a] = 0 // код cбой
			}
		}
	}
	return (false)
}

// яхонт4 - 3 расч статус источника питани¤ (неопр,норма,неиспр)
// ¬ паспорте должно быть:
// 0- неопределенность, 1-норма, 2-нет питани¤
//

func yahont42(a int, cod any) any {
	if (cod) == (0) { // код 0-неопр {
		dout[a] = 0
	} else {
		if (cod) == (3) { // код 3-норма {
			dout[a] = 1
		} else {
			if (cod) == (6) { // код 6-трев {
				dout[a] = 2
			} else {
				dout[a] = 0 // остальное cбой
			}
		}
	}
	return (false)
}

// -------------------------- яхонт16и -------------------------------
// —осто¤ние ѕќ∆ј–Ќџ’ шлейфов
// ¬ паспорте должно быть:
// 0- неопределенность, 1-норма, 2-пожар, 3-внимание
//

func yahont16_ps(a float32, cod string) any {
	if dost(cod) {
		if (val(cod) == (0)) || (val(cod) == (1)) || (val(cod) == (2)) { // 0-неопр// 1-кз 2-обрыв 6-обрыв {
			dout[int(a)] = 0 // ** неопр
		} else {
			if val(cod) == (3) { // 3-норма {
				dout[int(a)] = 1 // ** норма
			} else {
				if val(cod) == (4) { // 4-вним {
					dout[int(a)] = 3
				} else {
					if val(cod) == (5) { // 5-трев {
						dout[int(a)] = 2 // ** пожар/дым
					} else {
						dout[int(a)] = 0 // ** неисправн
					}
				}
			}
		}
	} else {
		dout[int(a)] = 0
	}
	return (false)
}

// —осто¤ние ќ’–јЌЌџ’ шлейфов
// ¬ паспорте должно быть:
// 0- неопределенность, 1-норма, 2-вскитые, 3- сн¤т с охраны
//

func yahont16_os(a int, cod any) any {
	if (cod) == (135) { // 0-неопр {
		dout[a] = 0
	} else {
		if (cod) == (132) { // 3-норма {
			dout[a] = 1
		} else {
			if (cod) == (134) { // 6-трев {
				dout[a] = 2
			} else {
				if ((cod) == (129)) || ((cod) == (130)) || ((cod) == (131)) {
					dout[a] = 3 // 6-сн¤т
				} else {
					dout[a] = 0 // cбой
				}
			}
		}
	}
	return (false)
}

// --- —осто¤ние источника питани¤ ---
// 0 - норма, 1- неисправность

func yahont16_pit(a float32, cod float32) {
	//2 переменные
	codInt := int(cod)
	dout[int(a)+0] = float32(ne(codInt&1, 0))
	dout[int(a)+1] = float32(ne(codInt&256, 0))
}

//bupg24_3.evl
// БУПГ24-3, БУПГ24-6
// Галеев 08.2019

//----------- <Адрес 11>. Состояние дискретных датчиков -------------------
//Единицы в битах регистра (кроме бита 10) означают, что соответствующие датчики  находятся в состоянии "Авария"
// нули - что соответствующие датчики находятся в состоянии "Норма".
//
// Соответствие битов регистра входным сигналам:
//   0 - неисправность датч Твх - БУПГ24-6
//   1 - не используются//
//   2 - сигнал перегрева//
//   3 - сигнал <Давление газа высокое>//
//   4 - сигнал <Давление газа низкое>//
//   5 - сигнал <Давление продукта высокое>//
//   6 - сигнал <Давление продукта низкое>//
//   7 - сигнал <Разрежение низкое>//
//   8 - сигнал <Уровень теплоносителя низкий>//
//   9 - сигнал <Прорыв газа>//
//   10 - сигнал <>// 1 - наличие пламени, 0 - его отсутствие.
//   11 - сигнал <Расход продукта низкий>//
//   12 - сигнал <Давление запальника высокое>//
//   13 - сигнал <Загазованность>//
//   14 - сигнал <Неисправность аналогового датчика  температуры теплоносителя>//
//   15 - сигнал <Неисправность аналогового датчика температуры газа на выходе>.
//-----------------------------
//--------<Адрес 12>. Аварийные состояния датчиков
//Соответствие битов регистра входным сигналам: аналогично регистрам 1, 11.
//При аварийном отключении подогревателя по какому-либо сигналу бит,
//соответствующий этому сигналу, устанавливается в 1.
//Содержимое регистра сохраняется до следующего аварийного отключения подогревателя или отключения питания БУПГ.

func bupg243(bp int, cod float32) int {
	codInt := int(cod)
	dout[bp+0] = float32(ne(codInt&4, 0))   // Перегрев
	dout[bp+1] = float32(ne(codInt&8, 0))   // Ргаза высокое
	dout[bp+2] = float32(ne(codInt&16, 0))  // Ргаза низкое
	dout[bp+3] = float32(ne(codInt&32, 0))  // Ртн высокое
	dout[bp+4] = float32(ne(codInt&64, 0))  // Ртн низкое
	dout[bp+5] = float32(ne(codInt&128, 0)) // Разряжение низкое
	codInt = codInt / 256
	dout[bp+6] = float32(ne(codInt&1, 0))    // Уровень тн низкий
	dout[bp+7] = float32(ne(codInt&2, 0))    // Прорыв газа
	dout[bp+8] = float32(ne(codInt&4, 0))    // Пламя 0-нет 1-есть
	dout[bp+9] = float32(ne(codInt&8, 0))    // Расход низкий
	dout[bp+10] = float32(ne(codInt&16, 0))  // Давление запальника высокое
	dout[bp+11] = float32(ne(codInt&32, 0))  // Загазованность
	dout[bp+12] = float32(ne(codInt&64, 0))  // Пламя отсутств(3М), Неиспр датч температуры тн
	dout[bp+13] = float32(ne(codInt&128, 0)) // Неиспр датч температуры(3М) вых газа
	return 0
}

func bupg243_klap(bp int, cod float32) int {
	codInt := int(cod)
	dout[bp+0] = float32(ne(codInt&2, 0))  // клапан запальника
	dout[bp+1] = float32(ne(codInt&4, 0))  // клапан отсекателя
	dout[bp+2] = float32(ne(codInt&8, 0))  // клапан б.горения
	dout[bp+3] = float32(ne(codInt&16, 0)) // сигнал аварии
	dout[bp+4] = float32(ne(codInt&32, 0)) // звук сигнала аварии
	dout[bp+5] = float32(ne(codInt&64, 0)) // клапан безопасности
	return 0
}

func bupg(bp int, mode float32, cod11, cod12 float32) int {
	modeInt := int(mode)
	if eq(modeInt, 6) == 1 { // Авария
		bupg243(bp, cod12)
	} else {
		bupg243(bp, cod11)
	}
	return 0
}

// sgoes.evl
func eq(a, b int) int {
	if a == b {
		return 1
	}
	return 0
}

// sgoes_avar возвращает 1 если авария, иначе 0
func sgoes_avar(cod int) int {
	x := ne(cod&1, 1) // авария
	return x
}

// sgoes_porog1 возвращает 1 если превышен порог 1, иначе 0
func sgoes_porog1(cod int) int {
	x := eq(cod&2, 2) // порог 1 превышен
	return x
}

// sgoes_porog2 возвращает 1 если превышен порог 2, иначе 0
func sgoes_porog2(cod int) int {
	x := eq(cod&4, 4) // порог 2 превышен
	return x
}

// sgoes возвращает 1 если авария, иначе 0 (для совместимости со старым кодом САУ)
func sgoes(cod int) int {
	x := ne(cod&1, 1) // авария
	return x
}

// BOM2.evl
func ne(a, b int) int {
	if a != b {
		return 1
	}
	return 0
}

func bomOven(bp int, cod1, cod2 float32) {
	cod1Int := int(cod1)
	cod2Int := int(cod2)

	dout[bp+0] = float32(ne(cod1Int&2, 0))  // Клапан заправки
	dout[bp+1] = float32(ne(cod1Int&4, 0))  // Клапан пульсатора
	dout[bp+2] = float32(ne(cod1Int&16, 0)) // Клапан сброса
	dout[bp+3] = float32(ne(cod1Int&32, 0)) // Проникновение в одоризатор
	//cod := cod1Int / 256
	dout[bp+4] = float32(ne(cod1Int&2, 0) + 2*ne(cod1Int&4, 0))   // Низкая/Высокая температура
	dout[bp+5] = float32(ne(cod1Int&8, 0))                        // Высокое давление в коллекторе
	dout[bp+6] = float32(ne(cod1Int&16, 0))                       // Высокий перепад давления
	dout[bp+7] = float32(ne(cod1Int&32, 0) + 2*ne(cod1Int&64, 0)) // Низкий/Высокий уровень в расходной емкости
	dout[bp+8] = float32(ne(cod1Int&128, 0))                      // Ошибка выдачи дозы

	dout[bp+9] = float32(ne(cod2Int&1, 0))   // Авария
	dout[bp+10] = float32(ne(cod2Int&2, 0))  // Пожар
	dout[bp+11] = float32(ne(cod2Int&4, 0))  // Обрыв датчика потока
	dout[bp+12] = float32(ne(cod2Int&8, 0))  // Неисправность датчика давления в коллекторе
	dout[bp+13] = float32(ne(cod2Int&16, 0)) // Неисправность датчика давления в емкости
	dout[bp+14] = float32(ne(cod2Int&32, 0)) // Неисправность сигнализатора уровня
	dout[bp+15] = float32(ne(cod2Int&64, 0)) // Неисправность датчика температуры
	//cod = cod2Int / 256
	dout[bp+16] = float32(ne(cod2Int&1, 0)) // РИП норма
	dout[bp+17] = float32(ne(cod2Int&2, 0)) // РИП батарея норма
	dout[bp+18] = float32(ne(cod2Int&8, 0)) // Пульсатор
}

//vbp.evl
// Galeev
// Расчет обьема газа при работе на байпасе
// при старте зонда обнулить переменные
// 5 переменных
//aout[ba+0] - время открытия
//aout[ba+1] - время закрытия
//aout[ba+2] - время работы на байпасе, час
//aout[ba+3] - объем газа на байпасе
//aout[ba+4] - время работы на байпасе, сек
// kr_bp - положение байпасного крана
// Vpsut - объем газа за прошлые сутки
// ba - базовый адрес

func Vbp(kr_bp string, Vpsut float32, ba float32) {
	if val(kr_bp) == (1) {
		//if dost(aout[int(ba)+1]) {
		if true {
			aout[int(ba)] = float32(time.Now().Unix()) // время открытия
			aout[int(ba)+1] = 0

		}
		aout[int(ba)+2] = (float32(time.Now().Unix()) - aout[int(ba)]) / 3600 // в часах
		aout[int(ba)+4] = (float32(time.Now().Unix()) - aout[int(ba)])        // в сек
		aout[int(ba)+3] = float32(aout[int(ba)+2]) * Vpsut / 24
	}

	//if (val(kr_bp) == (2)) && ((dost(aout[ba+1])) == (0)) {
	if (val(kr_bp) == (2)) && (0) == (0) {
		aout[int(ba)+1] = float32(time.Now().Unix())
		aout[int(ba)+2] = (aout[int(ba)+1] - aout[int(ba)]) / 3600
		aout[int(ba)+4] = (aout[int(ba)+1] - aout[int(ba)]) // в сек
	}
}

func dostacc(sum float32, v1 string) float32 {
	if dost(v1) {
		sum = sum + val(v1)
	}
	return (sum)
}

func pdostnearest(med float32, v1 string, v2 string) string {
	if dost(v1) {
		if (math.Abs(float64(med - val(v1)))) > (math.Abs(float64(med - val(v2)))) {
			return (v2)
		}
		return (v1)
	}
	return (v2)
}

//2is3.evl
// возвращает индекс ближайшего из двух к среднему

// val,i - значение и индекс измерения
// dostval,di - значение и индекс достоверного измерения
// если val достоверно, возвращает индекс ближайшего из двух к среднему
// иначе водвращает di
func idostnearest(med float32, v1 string, i1 float32, v2 string, i2 float32, ii float32) float32 {
	if dost(v1) && dost(v2) {
		if (math.Abs(float64(med - val(v1)))) > (math.Abs(float64(med-val(v2))) + 0.002) { //+0.002 - чтобы убрать дребезг {
			return (i2)
		} else {
			return (i1)
		}
	} else {
		return (ii)
	}
}

// расчет индекса датчика давления для регулятора на двух эр-04
// 0 - эр04-12, 1 - эр04-21, 2 - эр04-22, подача в том же порядке
func regp3i(p1 string, p2 string, p3 string, self float32) float32 {
	i := float32(0)
	c := float32(0)
	p := float32(0)

	if dost(p1) {
		c = c + 1
		p = p + val(p1)
	}

	if dost(p2) {
		i = 1
		c = c + 1
		p = p + val(p2)
	}

	if dost(p3) {
		i = 2
		c = c + 1
		p = p + val(p3)
	}

	if c != 0 {
		p = p / c
		i = idostnearest(p, p1, 0, p3, 2, i)
		i = idostnearest(p, p2, 1, p1, 0, i)
		i = idostnearest(p, p3, 2, p2, 1, i)
	} else {
		i = self
	}

	return (i)
}

// выбор значения Рвых для регулятора по трем ан.датчикам
// pself - текущее значение параметра
func regp3p(p1 string, p2 string, p3 string, self float32) float32 {
	psum := dostacc(0, p1)
	psum = dostacc(psum, p2)
	psum = dostacc(psum, p3)

	c := 0
	if dost(p1) {
		c = 1
	}
	p := (p1)

	if dost(p2) {
		c = c + 1
		p = (p2)
	}

	if dost(p3) {
		c = c + 1
		p = (p3)
	}

	if c != 0 {
		x := val(pdostnearest(psum/float32(c), p1, p))
		x = val(pdostnearest(psum/float32(c), p2, p))
		x = val(pdostnearest(psum/float32(c), p3, p))
		return x
	} else {
		return 0
	}
}

// ПС 2 из 3 меньше, если достоверен только 1 датчик, не вырабатывать
// mux - множитель типа 90%
// тратит 3 переменные
//

func ps2is3Lt(p1 string, p2 string, p3 string, mux float32, pzad float32, T float32, vi float32) bool {
	a1 := 0
	if (valTrackLt(p1, 0.01*mux*pzad, T, vi)) > 0 {
		a1 = 1
	}
	a2 := 0
	if (valTrackLt(p2, 0.01*mux*pzad, T, vi+1)) > 0 {
		a2 = 1
	}
	a3 := 0
	if (valTrackLt(p3, 0.01*mux*pzad, T, vi+2)) > 0 {
		a3 = 1
	}
	x1 := 0
	x2 := 0
	x3 := 0
	if dost(p1) {
		x1 = 1
	}
	if dost(p2) {
		x2 = 1
	}
	if dost(p3) {
		x3 = 1
	}
	if x1+x2+x3 >= 2 {
		return ((a1 + a2 + a3) >= (2))
	}
	return (false)
}

// // ПС 2 из 3 больше, если достоверен только 1 датчик, не вырабатывать
// mux - множитель типа 110%
// тратит 3 переменные
func ps2is3Gt(p1 string, p2 string, p3 string, mux float32, pzad float32, T float32, vi float32) bool {
	a1 := 0
	if (valTrackGt(p1, 0.01*mux*pzad, T, vi)) > 0 {
		a1 = 1
	}
	a2 := 0
	if (valTrackGt(p2, 0.01*mux*pzad, T, vi+1)) > 0 {
		a2 = 1
	}
	a3 := 0
	if (valTrackGt(p3, 0.01*mux*pzad, T, vi+2)) > 0 {
		a3 = 1
	}
	x1 := 0
	x2 := 0
	x3 := 0
	if dost(p1) {
		x1 = 1
	}
	if dost(p2) {
		x2 = 1
	}
	if dost(p3) {
		x3 = 1
	}
	if x1+x2+x3 >= 2 {
		return ((a1 + a2 + a3) >= (2))
		//return(a3)
	}
	return (false)
}

//uug.evl
// Расчет объемов QY (за прошлые сутки), QD (с начала суток), за прошлый месяц для sevc и SF

// ********** Функция вычисления qd, qy
// t      - время устройства на этом шаге
// vs_sys - vsum, параметр непрерывный расход (исходный для всех расчетов)
// qf_sys - fix, параметр, где фиксируется непрер накопленный расход
//
//	при смене контр часа (уст извне, чтобы хранить)
//
// qy_sys - sys параметра qy (уст извне, чтобы хранить)
// qd_ind - индекс расчетной переменной qd
// qmax   - максимальный возможный расход за сутки, больше - недост
// chour  - контрактный час
// aout[vi+0] - расход с начала суток
// aout[vi+1] - для слежения за изменением времени
func hour(t time.Time) int {
	return t.Hour()
}

// updQyQd updates qy and qd based on the given parameters.
func upd_qyqd(t1 float32, vsSys, qfSys, qySys float32, vi float32, qmax float32, chour float32) {
	// Check if contract hour has arrived
	t := time.Now()
	if hour(t) != hour(time.Unix(int64(aout[int(vi)+1]), 0)) && hour(t) == int(chour) {
		qy := vsSys - qfSys
		if qy < 0 || qy > qmax {
			qy = 0 // Set qy to 0 if it's out of valid range
		}
		qySys = qy                  // Update qySys with the current accumulated value
		qfSys = vsSys               // Store the current accumulated value
		time.Sleep(2 * time.Second) // Sleep for 2 seconds to simulate usage of the new value
	}
	aout[int(vi)] = float32(vsSys - qfSys) // Calculate consumption since the start of the day
	aout[int(vi)+1] = float32(t.Unix())    // Update the last time
}

// ********** Функция вычисления qm - расхода за месяц
// t      - время устройства на этом шаге
// vs_sys - vsum, параметр непрерывный расход (исходный для всех расчетов)
// qf_sys - fix, параметр, где фиксируется непрер накопленный расход
//
//	на начало месяца (уст извне, чтобы хранить)
//
// qm_sys - sys параметра qy (уст извне, чтобы хранить)
// vi     - индекс переменной для расчетов
// qmax   - максимальный возможный расход за месяц, больше - недост
// chour  - контрактный час
// dout[vi+0] - для управления расчетом
// aout[vi+1] - для слежения за изменением времени
//
// month extracts the month part from the given time.
func month(t time.Time) int {
	return int(t.Month())
}

// updQmes updates qm based on the given parameters.
func updQmes(t time.Time, vsSys, qfSys, qmSys *float32, vi int, qmax float32, chour int, dout, aout []float32) {
	// Check if a new month has started
	if month(t)-1 != month(time.Unix(int64(aout[vi+1]), 0))-1 {
		dout[vi] = 1 // Enable delayed calculation of qm
	}

	// Check if the calculation is enabled and the contract hour has arrived
	if dout[vi] == 1 && hour(t) != hour(time.Unix(int64(aout[vi+1]), 0)) && hour(t) == chour {
		qm := *vsSys - *qfSys
		if qm < 0 || qm > qmax {
			qm = 0 // Set qm to 0 if it's out of valid range
		}
		*qmSys = qm                 // Update qmSys with the current accumulated value
		*qfSys = *vsSys             // Store the current accumulated value
		time.Sleep(2 * time.Second) // Sleep for 2 seconds to simulate usage of the new value
		dout[vi] = 0                // Reset the calculation flag
	}

	aout[vi+1] = float32(t.Unix()) // Update the last time
}

// переменные
// 1,2 - upd_qyqd sd
// 3,4 - upd_qmes sd
// 5,6 - upd_qmes sf2et1
// 7,8 - upd_qmes sf2et2

// oninit(t)
//aout[2]=Reps["SVC ВРЕМЯ БЕЛ"].Value        // чтобы замечать смену суток
//aout[4]=Reps["SVC ВРЕМЯ БЕЛ"].Value        // чтобы замечать смену суток
//aout[6]=Reps["БЕЛБ SF1-TIME"].Value         // чтобы замечать смену суток
//aout[8]=Reps["БЕЛБ SF2-TIME"].Value         // чтобы замечать смену суток
//aout[10]=Reps["БЕЛБ SF3-TIME"].Value         // чтобы замечать смену суток

func convertToBool(val float32) bool {
	return val != 0
}

//#include "eval.lib\set.evl"

// 01.06.15
// для вставки #include "eval.lib\set.evl"

// управление при условии достоверности
func setex(sys string, value float32) bool {
	if dost(sys) == false {
		return (false)
	}
	PKG.UpdateVal(sys, value, true)
	return (true)
}

// setwex - аналог встроенной SET_WAIT
// однако, в случае не успеха
// производится дополнительные 1 попытки
// достигнуть заданного соcтояния
func setwex(parameter string, value float32, timeout float64) float32 {
	PKG.UpdateVal(parameter+" УПР", value, true)
	if value == 3 { ///Если открываем
		value = 1
	} else if value == 4 { ///Если закрываем
		value = 2
	}
	if set_wait(parameter, value, timeout) != 0 {
		time.Sleep(250 * time.Millisecond)
		return set_wait(parameter, value, timeout)
	}
	return 0
}

// impuls
func impuls(sys string, t float64) float32 {
	x := set_wait(sys, 1, t)
	//time.Sleep((2*18) * time.Second)
	x = set_wait(sys, 0, t)
	return (x)
}

// установка значения с заданной чувствительностью
// возврат 1-установлено
// -
//
//	0-без реакции
func setSens(sys string, value float32, sens float32) bool {
	x := false
	if (math.Abs(float64(val(sys) - value))) > float64(sens) {
		x = setex(sys, value)
	}
	return x
}

func setwex_dost(sys string, value float32, timeout float64) any {
	if !dost(sys) {
		return (false)
	}
	return (set_wait(sys, value, timeout))
}

// #include "eval.lib\front.evl"
// front 0-> ne 0
// src - дискр сигнал
// id - номер переменной слежения
func front(src string, id float32) float32 {
	var x float32
	x = 0
	if dost(src) && val(src) != (dout[int(id)]) && val(src) != 0 {
		x = 1
	}
	dout[int(id)] = val(src)
	return (x)
}

// Тест КРБП
// Проверено в Телепаново
// Галеев 19.03.15

// 25.06.2015 :Галеев. Проверено в Таптыково
// 1.добавлена обработка события когда при начале теста положение крбп
//   сильно отличается от задания. это сразу дает неисправность
// 2.исключена ошибочная засылка задания выше 100%

//#INCLUDE "eval.lib\klap_test.evl"

//u=klap_test(u_1,Reps["РУЧПОЛБП ЯНАУ"].Value,Reps["ПОЛОЖЗАДВ ЯНЛ"].Value,33,Reps["КР БП ЯНАУ"].Value,7,Reps["РЕЖИМ ГРС"].Value)

// час текущего времени сау
func curhour() float32 {
	curtime := time.Now()
	return (float32(curtime.Hour()))
}

// тест клапана
// man - ручное задание
// pol - положение клапана
// u - сигнал управления из pid
// dout[vi+0] - пс неисправности клапана
// aout[vi+1] - отсчет времени теста
// aout[vi+2] - предыдущий час
// dout[vi+3] - внеочередная проверка
func klap_test(u float32, man float32, pol float32, vi float32, bp_kr any, t any, rejim_grs any) float32 {
	if (bp_kr) == (2) {

		if (aout[int(vi+1.0)]) == (0) {
			u = float32(man) // без проверки было бы так и все
		}

		h := curhour()
		if (rejim_grs) != (0) {
			a := false
			if dout[int(vi+3)] > 0 {
				a = true
			}
			if (h) != (aout[int(vi+2)]) && ((h) == (t)) && ((aout[int(vi+1)]) == (0)) || a { // dout - внеочередной тест {
				if (math.Abs(float64(pol - man))) < (8) {
					aout[int(vi+1)] = GETTICKS(0) // ждем
				} else {
					dout[int(vi)] = 1
				}
				dout[int(vi+3)] = 0 // сбросить флаг внеочередного теста
			}

			if (aout[int(vi+1)]) != (0) {

				u = float32(math.Min(float64(man)+15, 100)) // все время теста держим задание

				if ((GETTICKS(aout[int(vi+1)]) * TICKSIZE()) >= (40)) || ((pol) > float32(math.Min(99, float64(man+8)))) {

					if (pol) < (man + 8) {
						dout[int(vi)] = 1
					} else {
						dout[int(vi)] = 0
					}
					aout[int(vi+1)] = 0 // сам приедет обратно
				}
			}
		}

		aout[int(vi+2)] = h
	}
	return (u)
}

// regim.evl
// переходы в режим по кнопкам или командам
// vi_mode - номер перем режима грс (уст извне, 0-по месту, 1-пу, 2-арм)
// evt - команда/кнопка
// vi  - номер переменной слежения
// v1,v2 - значения режима грс, между которыми переход
func hev(vi_mode float32, vi float32, evt string, v1 float32, v2 float32) {
	if front(evt, vi) == 1 { // нажата/подана {
		if dout[int(vi_mode)] == (v1) {
			dout[int(vi_mode)] = v2 // туды
		} else {
			if (dout[int(vi_mode)]) == (v2) {
				dout[int(vi_mode)] = v1 // сюды
			}
		}
	}
}

// реакция на команды и кнопки перехода в режим
// по кнопкам или командам 1 переходы 0-2-0
// по кнопкам или командам 2 переходы 1-2-1
// переходы 0-1-0 запрещены
// cmd1,cmd2 - команды пользователя (упр извне вычислитель)
// but1,but2 - кнопки смены режима (д.вх)
// vi - начальный номер области переменных
// vi+0 - номер перем режима грс (уст извне, 0-по месту, 1-пу, 2-арм)
// vi+1..vi+4 - слежение за ком-кнопками
// vi+5, vi+6 - задержка при восст команд реж ту грс
// vi+7, vi+8 - тела команд реж ту грс
func modes(vi float32, cmd1 string, cmd2 string, but1 string, but2 string) {
	hev(vi+0, vi+1, cmd1, 0, 2)
	hev(vi+0, vi+2, but1, 0, 2)
	hev(vi+0, vi+3, cmd2, 1, 2)
	hev(vi+0, vi+4, but2, 1, 2)
	if valTrack(val(cmd1), 5, float32(vi+5)) == 1 {
		dout[int(vi)+7] = 0
	}
	if (valTrack(val(cmd2), 5, float32(vi+6))) == 1 {
		dout[int(vi)+8] = 0
	}
	//return(false)
}

// переходы в режим по команде от алгоритма
// vi_mode - номер перем режима грс (уст извне, 0-по месту, 1-пу, 2-арм)
// cmd - команда, значение режима грс, куда надо перевести
// vi  - номер переменной слежения
// vi+1 - тело команды
func cmdmode_in(vi_mode float32, vi float32, cmd string) {
	if dost(cmd) && val(cmd) != (dout[int(vi)]) {
		dout[int(vi_mode)] = val(cmd)
		dout[int(vi)+1] = dout[int(vi_mode)]
	}
	dout[int(vi)] = val(cmd)
	dout[int(vi)+1] = dout[int(vi_mode)]
	//return(false)
}

func to_mest(vi_mode float32, vi float32) float32 {
	if (dout[int(vi)]) == (1) {
		dout[int(vi_mode)] = 0 // по месту
		time.Sleep((3) * time.Second)
		dout[int(vi)] = 0 // взвод
	}
	return (0)
}

// Библиотека для работы с одоризаторами БОМ
// #include "eval.lib\BOM.evl"

// vi    - номер свободной выходной переменной для таймеров  Требуется 2 шт.
// q1    - мгновенный расход газа по прибору учета газа
// qz    - расход газа замещающий, при недостоверных
//	  	  данных с прибора учета газа (суперфло)
// mode  - режим одоризации
//	  	  для БОМ:  	0-автоматич от суперфло (посредством реле)
//		    		1-автоматич от САУ (требуется засылка)
//		    		2-ручное задание расхода газа
// cnt_sys - #сист номер расхода газа одоризатора в который засылать
// Т 	- период засылки.
// ПРИМЕР ИСПОЛЬЗОВАНИЯ:
// в ините
//  aout[16]=GETTICKS(0)
//  aout[17]=GETTICKS(0)
// в тексте
//  x=setq_periodic(16,(Reps["МГН РАС-1 ДЮРТ"].Value+Reps["МГН РАС-2 ДЮРТ"].Value),Reps["QМГН ЗАМ ДЮРТ"].Value,Reps["РЕЖ ОДОР ДЮРТ"].Value,Reps["QГ ОДОР ДЮРТ"].sys_num,30)
// или
//  x=setq_one((Reps["ПР СУТ-1 ДЮРТ"].Value+Reps["ПР СУТ-2 ДЮРТ"].Value)/24,Reps["QМГН ЗАМ ДЮРТ"].Value,Reps["QГ ОДОР ДЮРТ"].sys_num)

func setq(q float32, cnt_sys string) {
	if dost(cnt_sys) {
		if (math.Abs(float64(val(cnt_sys) - q))) > (5) { // если расход мало изменился не засылаем {
			PKG.UpdateVal(cnt_sys, q, true)

		}
	}
}

func setq_one(q string, qz float32, cnt_sys string) {
	if dost(q) {
		setq(val(q), cnt_sys)
	} else {
		setq(qz, cnt_sys)
	}
}

func setq_periodic(vi float32, q1 string, qz float32, mode any, cnt_sys string, T float32) float32 {
	if (mode) != (2) { // не автомат расход {
		return (0)
	}

	if (GETTICKS(aout[int(vi)]) * TICKSIZE()) >= (T) {
		if dost(q1) {
			setq(val(q1), cnt_sys)
		} else {
			var k float32
			k = 0
			if !dost(q1) {
				k = 1
			}
			if valTrack(k, 60, float32(vi+1)) > 0 { // если расход недост ждем 60 сек потом засылаем замещенный {
				setq(qz, cnt_sys)
			}
		}
		aout[int(vi)] = GETTICKS(0)
	}

	return (0)
}

func val(Raper string) float32 {
	PKG.InputMap.Mu.Lock()
	value := PKG.InputMap.Out[Raper].Value
	PKG.InputMap.Mu.Unlock()
	return value
}

//sost_grs.evl
// определение состояния ГРС по кранам

func sost_grs() float32 {
	var nitka1, nitka2 float32
	if (val("КР ЛРЕД1 КРАС")) != (2) {
		nitka1 = 1
	}
	if (val("КР ЛРЕД2 КРАС")) != (2) {
		nitka2 = 1
	}

	imp_on := 0 //val("КОМ ИМП Р ЮМ В")  &&  val("АЛГ ИМП Р ЮМ В")

	red_norma := (((nitka1) == (1)) && ((nitka2) == (1)))
	red_zakr := (((nitka1) == (0)) && ((nitka2) == (0)))
	red_1 := !red_zakr && (((nitka1) == (0)) || ((nitka2) == (0)))

	pg_norma := ((val("КР ПГВХ КРАС")) != (2) && (val("КР ПГВЫХ КРАС")) != (2)) && ((val("КРАН БППГ КРАС")) == (2))
	pg_bp := (((val("КР ПГВХ КРАС")) == (2)) || ((val("КР ПГВЫХ КРАС")) == (2))) && ((val("КРАН БППГ КРАС")) == (1))
	pg_zakr := (((val("КР ПГВХ КРАС")) == (2)) || ((val("КР ПГВЫХ КРАС")) == (2))) && ((val("КРАН БППГ КРАС")) == (2))

	var x float32
	var t bool
	if imp_on > 0 {
		t = true
	} else {
		t = false
	}
	if (val("КРАН ОХР КРАС")) != (2) && (val("КРАН ВХОД КРАС")) != (2) && (val("КРАН ВЫХ КРАС")) != (2) && ((val("КРАН БАЙП КРАС")) == (2)) && red_norma && pg_norma {
		//x = 0// работа
	} else {
		if (val("КРАН ОХР КРАС")) == (2) {
			x = 1 // отключена от МГ
		} else {
			if (val("КРАН БАЙП КРАС")) != (2) {
				x = 3 // на бп
			} else {
				if ((val("КРАН ВХОД КРАС")) == (2)) || ((val("КРАН ВЫХ КРАС")) == (2)) || (red_zakr && !t) || pg_zakr {
					x = 2 // остановлена
				} else {
					if red_1 {
						x = 4 // работа с отключенной лр
					} else {
						if red_zakr && t {
							x = 6 // импульсный режим
						} else {
							if pg_bp {
								x = 5 // работа с отключ пг
							}
						}
					}
				}
			}
		}
	}
	return (x)
}
