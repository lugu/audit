package fuzz

import (
	"github.com/lugu/qiloop/bus/client"
	"github.com/lugu/qiloop/type/value"
)

type CapabilityMap client.CapabilityMap

func (cm CapabilityMap) WithBasics() {
	cm["ClientServerSocket"] = value.Bool(true)
	cm["MessageFlags"] = value.Bool(true)
	cm["MetaObjectCache"] = value.Bool(true)
	cm["RemoteCancelableCalls"] = value.Bool(true)
}

func (cm CapabilityMap) WithExtras() {
	cm["Another field 1"] = value.Bool(true)
	cm["Another field 2"] = value.Bool(false)
	cm["Another field 3"] = value.Int8(0)
	cm["Another field 4"] = value.Int8(-42)
	cm["Another field 5"] = value.Uint8(0)
	cm["Another field 6"] = value.Uint8(42)
	cm["Another field 7"] = value.Int16(0)
	cm["Another field 8"] = value.Int16(-42)
	cm["Another field 9"] = value.Uint16(0)
	cm["Another field 10"] = value.Uint16(42)
	cm["Another field 11"] = value.Int(0)
	cm["Another field 12"] = value.Int(42)
	cm["Another field 13"] = value.Long(0)
	cm["Another field 14"] = value.Long(42 << 42)
	cm["Another field 15"] = value.Float(-1.234)
	cm["Another field 16"] = value.Float(0)
	cm["Another field 17"] = value.String("")
	cm["Another field 18"] = value.String("keep testing")
}

func (cm CapabilityMap) WithStrings() {
	// http://clagnut.com/blog/2380/
	cm["Arabic"] = value.String("صِف خَلقَ خَودِ كَمِثلِ الشَمسِ إِذ بَزَغَت — يَحظى الضَجيعُ بِها نَجلاءَ مِعطا")
	cm["Bulgarian"] = value.String("х чудна българска земьо, полюшвай цъфтящи жита.")
	cm["Chinese"] = value.String("視野無限廣，窗外有藍天")
	cm["German"] = value.String("Victor jagt zwölf Boxkämpfer quer über den großen Sylter Deich")
	cm["Greek"] = value.String("Ταχίστη αλώπηξ βαφής ψημένη γη, δρασκελίζει υπέρ νωθρού κυνός Takhístè alôpèx vaphês psèménè gè, draskelízei ypér nòthroý kynós")
	cm["Hebrew"] = value.String("דג סקרן שט בים מאוכזב ולפתע מצא חברה dg sqrn šṭ bjM mʾwkzb wlptʿ mṣʾ ḥbrh")
	cm["Hindi"] = value.String("ऋषियों को सताने वाले दुष्ट राक्षसों के राजा रावण का सर्वनाश करने वाले विष्णुवतार भगवान श्रीराम, अयोध्या के महाराज दशरथ के बड़े सपुत्र थे।")
	cm["Icelandic"] = value.String("Þú dazt á hnéð í vök og yfir blóm sexý pæju.")
	cm["Japanese"] = value.String("いろはにほへと ちりぬるを わかよたれそ つねならむ うゐのおくやま けふこえて あさきゆめみし ゑひもせす（ん）")
	cm["Korean"] = value.String("키스의 고유조건은 입술끼리 만나야 하고 특별한 기술은 필요치 않다.")
	cm["Sanskrit"] = value.String("कः खगौघाङचिच्छौजा झाञ्ज्ञोऽटौठीडडण्ढणः। तथोदधीन् पफर्बाभीर्मयोऽरिल्वाशिषां सहः।।")
	cm["Thai"] = value.String("เป็นมนุษย์สุดประเสริฐเลิศคุณค่า กว่าบรรดาฝูงสัตว์เดรัจฉาน จงฝ่าฟันพัฒนาวิชาการ อย่าล้างผลาญฤๅเข่นฆ่าบีฑาใคร ไม่ถือ")
}

func (cm CapabilityMap) WithCredential(user, token string) {
	cm[client.KeyUser] = value.String(user)
	cm[client.KeyToken] = value.String(token)
}

func basicNao() client.CapabilityMap {
	cm := CapabilityMap{}
	cm.WithBasics()
	cm.WithCredential("nao", "nao")
	return client.CapabilityMap(cm)
}

func justNao() client.CapabilityMap {
	cm := CapabilityMap{}
	cm.WithCredential("nao", "nao")
	return client.CapabilityMap(cm)
}

func extraNao() client.CapabilityMap {
	cm := CapabilityMap{}
	cm.WithExtras()
	cm.WithCredential("nao", "nao")
	return client.CapabilityMap(cm)
}

func stringsNao() client.CapabilityMap {
	cm := CapabilityMap{}
	cm.WithStrings()
	cm.WithCredential("nao", "nao")
	return client.CapabilityMap(cm)
}

func GetSamples() []client.CapabilityMap {
	samples := make([]client.CapabilityMap, 0)
	samples = append(samples, basicNao())
	samples = append(samples, justNao())
	samples = append(samples, extraNao())
	samples = append(samples, stringsNao())
	return samples
}
