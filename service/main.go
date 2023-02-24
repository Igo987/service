package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/antondzhukov/skillbox-diploma/service/models"
	"github.com/gorilla/mux"
)

// список допустимых провайдеров
var providersSMS = []string{"Topolo", "Rond", "Kildy"}
var providersVoiceCall = []string{"TransparentCalls", "E-Voice", "JustPhone"}
var providersEmail = []string{"Gmail", "Yahoo", "Hotmail", "MSN", "Orange", "Comcast", "AOL", "Live", "RediffMail", "GMX", "Proton Mail", "Yandex", "Mail.ru"}

// адреса
var pathVoice = "../simulator/voice.data"
var pathEmail = "../simulator/email.data"
var pathSMS = "../simulator/sms.data"
var pathBiling = "../simulator/billing.data"
var pathCountry = "../simulator/country.json"

// https://gist.github.com/adamveld12/c0d9f0d5f0e1fba1e551 - аббревиатуры пишутся CAPS`ом
const URL = "http://127.0.0.1:8383/mms"
const URLSupport = "http://127.0.0.1:8383/support"
const URLAccendent = "http://127.0.0.1:8383/accendent"

// getResultData сбор всх данных
func getResultData() (models.ResultSetT, error) {
	sms, err := createSMSDataList(pathSMS, providersSMS, countriesCode)
	filterSMS, err := filterFromAplha2(sms, err, parseCountryCode)
	resultSMS := getSMSResultData(filterSMS)
	mms, err := getDataMMS(URL, providersSMS, countriesCode)
	filterMMS, err := filterMMSFromAplha2(mms, err, parseCountryCode)
	resultMMS := getMMSresultData(filterMMS)
	voiceCall, err := createVoice(pathVoice, providersVoiceCall)
	resultVoiceCall, err := filterVoiceFromAplha2(voiceCall, err, countriesCode)
	email, err := createEmailData(pathEmail, providersEmail)
	filterEmail, err := filterEmailFromAplha2(email, err, countriesCode)
	resultEmail := getResultEmail(filterEmail)
	support, err := getDataSupport(URLSupport)
	resultSupport := getSupport(support)
	resultBiling, err := getBilingData(pathBiling)
	resultDataIcident, err := getIncidentData(URLAccendent)
	if err != nil {
		return models.ResultSetT{}, err
	}
	var res models.ResultSetT
	res.SMS = resultSMS
	res.MMS = resultMMS
	res.VoiceCall = resultVoiceCall
	res.Billing = resultBiling
	res.Incidents = resultDataIcident
	res.Support = resultSupport
	res.Email = resultEmail
	return res, nil
}

// Функция filterFromAplha2 для SMS для проверки соответствия значения списку из Aplha2
func filterFromAplha2(res []models.SMSData, e error, m map[string]string) ([]models.SMSData, error) {
	if e == nil {
		for i, v := range res {
			_, ok := m[v.Country]
			if !ok {
				res = append(res[:i], res[i+1:]...)
			}
		}
		return res, nil
	} else {
		return res, e
	}

}

// Функция filterMMSFromAplha2 для MMS для проверки соответствия значения списку из Aplha2
func filterMMSFromAplha2(res []models.MMSData, e error, m map[string]string) ([]models.MMSData, error) {
	if e == nil {
		for i, v := range res {
			_, ok := m[v.Country]
			if !ok {
				res = append(res[:i], res[i+1:]...)
			}
		}
		return res, nil
	} else {
		return res, e
	}
}

// Функция для проверки соответствия значения списку из Aplha2
func filterVoiceFromAplha2(res []models.VoiceCall, e error, m map[string]string) ([]models.VoiceCall, error) {
	if e == nil {
		for i, v := range res {
			_, ok := m[v.Country]
			if !ok {
				res = append(res[:i], res[i+1:]...)
			}
		}
		return res, nil
	} else {
		return res, e
	}
}

// Функция filterEmailFromAplha2 для проверки соответствия значения списку из Aplha2
func filterEmailFromAplha2(res []models.EmailData, e error, m map[string]string) ([]models.EmailData, error) {
	if e == nil {
		for i, v := range res {
			_, ok := m[v.Country]
			if !ok {
				res = append(res[:i], res[i+1:]...)
			}
		}
		return res, nil
	} else {
		return res, e
	}
}

// Функция filterFromProviders для проверки соответствия значения списку провайдеров
func filterFromProviders(val string, arr []string) bool { // проверяю на наличие в списке допустимых провайдеров
	find := false
	for _, v := range arr {
		if val == v {
			find = true
		}
	}
	return find
}

// читать SMS
func createSMSDataList(path string, arr []string, countriesCode map[string]string) ([]models.SMSData, error) {
	SMSDataList := make([]models.SMSData, 0)
	file, err := os.Open(path)
	if err != nil {
		log.Print(err)
		return SMSDataList, err
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	if err != nil {
		log.Print(err)
		return SMSDataList, err
	}
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Print(err)
		return SMSDataList, err
	}
	for _, line := range data {

		var rec models.SMSData
		for a := 0; a < len(line); a++ {
			for item := range line {
				res := strings.Split(line[item], ";")
				// Строки, в которых меньше четырёх полей данных не допускаются.
				if len(res) != 4 {
					continue
				} else {
					// В результат допускаются только сообщения с провайдерами из списка
					if filterFromProviders(res[3], arr) {
						rec.Country = countriesCode[res[0]]
						rec.Bandwidth = res[1]
						rec.ResponseTime = res[2]
						rec.Provider = res[3]
						SMSDataList = append(SMSDataList, rec)
					} else {
						continue
					}
				}
			}
		}

	}
	// сортировка данных по полю Provider
	sort.Slice(SMSDataList, func(i, j int) bool { return SMSDataList[i].Provider < SMSDataList[j].Provider })
	return SMSDataList, nil
}

// getSMSResultData приводиим данные об SMS к типу [][]models.SMSData
func getSMSResultData(sms []models.SMSData) [][]models.SMSData {
	smsData := make([][]models.SMSData, 0)

	smsData = append(smsData, sms)
	doubleSms := make([]models.SMSData, len(sms))
	copy(doubleSms, sms)
	sort.Slice(doubleSms, func(i, j int) bool { return doubleSms[i].Country < doubleSms[j].Country })
	smsData = append(smsData, doubleSms)
	return smsData
}

// функция getMMSresultData приводиим данные об MMS к типу [][]models.MMSData
func getMMSresultData(sms []models.MMSData) [][]models.MMSData {
	smsData := make([][]models.MMSData, 0)
	smsData = append(smsData, sms)
	doubleSms := make([]models.MMSData, len(sms))
	copy(doubleSms, sms)
	sort.Slice(doubleSms, func(i, j int) bool { return doubleSms[i].Country < doubleSms[j].Country })
	smsData = append(smsData, doubleSms)
	return smsData

}

// функция createVoice производит сбор данных о системе VoiceCall
func createVoice(url string, arr []string) ([]models.VoiceCall, error) {
	var data []models.VoiceCall
	file, err := os.Open(url)
	if err != nil {
		log.Println(err)
		return data, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = ';'

	for {
		var item models.VoiceCall
		record, e := reader.Read()
		if e != nil {
			if e == io.EOF {
				err = nil
				break
			}

		}
		if len(record) != 8 {
			continue
		} else {
			if filterFromProviders(record[3], arr) {
				for i := 0; i <= 7; i++ {
					item.Country = record[0]
					item.Bandwidth = record[1]
					item.ResponseTime, _ = strconv.Atoi(record[2])
					item.Provider = record[3]
					value, _ := strconv.ParseFloat(record[4], 32)
					item.ConnectionStability = float32(value)
					item.TTFB, _ = strconv.Atoi(record[5])
					item.VoicePurity, _ = strconv.Atoi(record[6])
					item.MedianOfCallsTime, _ = strconv.Atoi(record[7])
				}
			}

			data = append(data, item)
		}
	}
	return data, nil
}

// функция getBilingData производит сбор данных о системе Billing
func getBilingData(path string) (models.BillingData, error) {
	var res models.BillingData
	var summ uint8 // Сумма степеней каждого бита должна быть присвоена переменной с типом uint8.
	text, err := os.ReadFile(path)

	if err != nil {
		log.Fatal(err)
		return res, err
	}

	text_res := strings.Split(string(text), "")
	text_res_bool := make([]bool, 0)
	count := -1 // счётчик степени. Чтобы начать с нуля.
	for i := len(text_res) - 1; i >= 0; i-- {
		item, _ := strconv.Atoi(text_res[i])
		count++
		if item != 0 {
			summ += uint8(math.Pow(2, float64(count))) // перевод из двоичной в десятичную системы
			text_res_bool = append(text_res_bool, true)
		} else {
			text_res_bool = append(text_res_bool, false)
		}

	}
	if len(text_res_bool) == 6 {
		res.CreateCustomer = text_res_bool[0]
		res.Purchase = text_res_bool[1]
		res.Payout = text_res_bool[2]
		res.Recurring = text_res_bool[3]
		res.FraudControl = text_res_bool[4]
		res.CheckoutPage = text_res_bool[5]
	}

	return res, nil
}

// функция createEmailData производит сбор данных о системе Email
func createEmailData(url string, arr []string) ([]models.EmailData, error) {
	var data []models.EmailData
	file, err := os.Open(url)
	if err != nil {
		log.Println("email", err)
		return data, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = ';'

	for {
		var item models.EmailData
		record, e := reader.Read()
		if e != nil {
			if e == io.EOF {
				err = nil
				break
			}
		}
		if len(record) != 3 {
			continue
		} else {
			if filterFromProviders(record[1], arr) {
				for i := 0; i <= 2; i++ {
					item.Country = record[0]
					item.Provider = record[1]
					value, _ := strconv.Atoi(record[2])
					item.DeliveryTime = value
				}
				data = append(data, item)
			}
		}
	}
	return data, nil
}

// функция getResultEmail приводит данные о Email к типу models.ResultEmail
func getResultEmail(e []models.EmailData) models.ResultEmail {
	res := make(map[string][]models.EmailData)
	result := make(models.ResultEmail)
	fast := make([]models.EmailData, 0)
	slow := make([]models.EmailData, 0)
	filterCountry := make(map[string]string)
	for _, v := range e {
		filterCountry[v.Country] = v.Provider
	}
	uniqCountry := make([]string, 0)
	for key := range filterCountry {
		uniqCountry = append(uniqCountry, key)

	}
	for _, v := range e {
		for _, i := range uniqCountry {
			if v.Country == i {
				res[v.Country] = append(res[v.Country], v)
			}
		}
	}
	for key := range res {
		for _, i := range uniqCountry {
			if key == i {
				sort.Slice(res[key], func(m, v int) bool { return res[key][m].DeliveryTime > res[key][v].DeliveryTime })
			}
		}

		fast = append(res[key][:3])
		slow = append(res[key][(len(res[key]) - 3):])
		result[key] = append(result[key], fast)
		result[key] = append(result[key], slow)

	}
	return result
}

// функция getDataMMS производит сбор данных о системе MMS
func getDataMMS(url string, arr []string, countriesCode map[string]string) ([]models.MMSData, error) {
	var dataArray []models.MMSData
	req, err := http.Get(url)
	if err != nil {
		log.Println("MMS", err)
		return dataArray, err
	}
	if req.StatusCode == 200 {
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Println("MMS", err)
		}
		err = json.Unmarshal(data, &dataArray)
		if err != nil {
			return dataArray, err
		}
		for i, item := range dataArray {

			if filterFromProviders(item.Provider, arr) {
				continue
			} else {
				dataArray = append(dataArray[:i], dataArray[i+1:]...)
			}
		}
	} else {
		log.Printf("MMS. Status: %v", req.StatusCode)
	}
	defer req.Body.Close()
	sort.Slice(dataArray, func(i, j int) bool { return dataArray[i].Provider < dataArray[j].Provider })
	parse := make([]models.MMSData, 0)
	for _, v := range dataArray {
		v.Country = countriesCode[v.Country]
		parse = append(parse, v)

	}
	return parse, nil
}

// функция getDataSupport производит сбор данных о системе Support
func getDataSupport(url string) ([]models.SupportData, error) {
	var dataArray []models.SupportData
	req, err := http.Get(url)
	if err != nil {
		fmt.Printf("Произошла ошибка - %s", err)
		return dataArray, err
	}
	defer req.Body.Close()
	if req.StatusCode == 200 {
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Printf("Произошла ошибка - %s", err)
		}
		err = json.Unmarshal(data, &dataArray)
		if err != nil {
			log.Printf("Произошла ошибка - %s", err)
		}
		for _, item := range dataArray {
			dataArray = append(dataArray, item)
		}
	} else {
		log.Printf("Support. Status: %v", req.StatusCode)
		return dataArray, err
	}
	return dataArray, nil
}

// функция getIncidentData производит сбор данных о системе Incident
func getIncidentData(url string) ([]models.IncidentData, error) {
	var dataArray []models.IncidentData
	req, err := http.Get(url)
	if err != nil {
		fmt.Printf("Произошла ошибка - %s", err)
		return dataArray, err
	}
	defer req.Body.Close()
	if req.StatusCode == 200 {
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Printf("Произошла ошибка - %s", err)
		}
		err = json.Unmarshal(data, &dataArray)
		if err != nil {
			log.Printf("Произошла ошибка - %s", err)
		}
		for _, item := range dataArray {
			if item.Status != "closed" {
				continue
			} else if item.Status != "active" {
				continue
			}

			dataArray = append(dataArray, item)
		}
	} else {
		log.Printf("Error. Status: %v", req.StatusCode)
		return dataArray, nil
	}
	if len(dataArray) > 0 {
		sort.Slice(dataArray, func(i, j int) bool { return dataArray[i].Status < dataArray[j].Status })
	}

	return dataArray, nil
}

// handleConnection
func handleConnection(w http.ResponseWriter, r *http.Request) {
	var result models.ResultT
	if resultSetT, err := getResultData(); err == nil {
		result = models.ResultT{
			Status: true,
			Data:   &resultSetT,
			Error:  "",
		}
	} else {
		result = models.ResultT{
			Status: false,
			Data:   nil,
			Error:  "Error",
		}

	}
	response, _ := json.Marshal(result)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Write(response)
}

// функция getSupport приводит данные о SupportData к типу []int
func getSupport(s []models.SupportData) []int {
	supportWorkLoad := make([]int, 0)
	var summ int
	var workloadLevel int
	minutesPerTask := 60 / 18
	for _, v := range s {
		summ += v.ActiveTickets
	}
	minutesPerNextTask := (minutesPerTask * summ)
	if (summ) >= 16 {
		workloadLevel = 3
	}
	if (summ <= 9) && (summ < 16) {
		workloadLevel = 2
	}
	if summ < 9 {
		workloadLevel = 1
	}
	supportWorkLoad = append(supportWorkLoad, workloadLevel)
	supportWorkLoad = append(supportWorkLoad, minutesPerNextTask)
	return supportWorkLoad
}

//countriesCode map для Aplha2
var countriesCode = make(map[string]string)
var parseCountryCode = make(map[string]string)

func startServ() {
	r := mux.NewRouter()
	r.HandleFunc("/", handleConnection)
	log.Fatal(http.ListenAndServe(":8282", r))

}

func main() {
	// Читаем данные с файла country.json => делаем из них map countriesCode
	jsonFile, err := os.Open(pathCountry)
	if err != nil {
		log.Println(err)
	}

	defer jsonFile.Close()

	data_country, err := ioutil.ReadAll(jsonFile)

	if err != nil {
		log.Println("Error reading json data:", err)
	}
	var сountries []models.Country

	er := json.Unmarshal(data_country, &сountries)
	if er != nil {
		log.Println("Error unmarshalling json data:", err)
	}
	// задаём соответствие вида "код":"страна"
	for i := 0; i < len(сountries); i++ {
		countriesCode[сountries[i].Country] = сountries[i].Code
	}
	// countriesCode задаём обратное соответствие чтобы получить полное название страны по её коду
	for key, value := range countriesCode {
		parseCountryCode[value] = key
	}
	go startServ()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)
	go func() {
		<-sigs
		done <- true
	}()
	<-done
	dev, _ := getResultData()
	fmt.Println(dev.VoiceCall)
}
