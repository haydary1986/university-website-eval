package database

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"website-eval-system/models"

	"golang.org/x/crypto/bcrypt"
)

func Seed() {
	// Check if reset is requested
	if os.Getenv("RESET_DB") == "true" {
		log.Println("RESET_DB=true, dropping all data and re-seeding...")
		DB.Exec("DELETE FROM active_sessions")
		DB.Exec("DELETE FROM login_attempts")
		DB.Exec("DELETE FROM audit_logs")
		DB.Exec("DELETE FROM blocked_ips")
		DB.Exec("DELETE FROM reviews")
		DB.Exec("DELETE FROM submission_items")
		DB.Exec("DELETE FROM submissions")
		DB.Exec("DELETE FROM criteria")
		DB.Exec("DELETE FROM categories")
		DB.Exec("DELETE FROM users")
		DB.Exec("DELETE FROM universities")
		DB.Exec("DELETE FROM academic_years")
	}

	// Check if already seeded
	var count int64
	DB.Model(&models.User{}).Count(&count)
	if count > 0 {
		log.Println("Database already seeded, skipping...")
		return
	}

	log.Println("Seeding database...")

	seedCategories()
	seedUniversities()
	seedAcademicYears()
	seedSuperAdmin()
	seedUniversityUsers()

	log.Println("Database seeded successfully!")
}

func seedSuperAdmin() {
	hash, _ := bcrypt.GenerateFromPassword([]byte("Sakina1990"), bcrypt.DefaultCost)
	admin := models.User{
		Username: "haydary1986",
		Password: string(hash),
		FullName: "مدير النظام",
		Email:    "admin@mohe.gov.iq",
		Role:     "super_admin",
	}
	DB.Create(&admin)
}

func seedAcademicYears() {
	years := []models.AcademicYear{
		{
			Name:      "2024-2025",
			StartDate: time.Date(2024, 9, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2025, 8, 31, 0, 0, 0, 0, time.UTC),
			IsActive:  false,
		},
		{
			Name:      "2025-2026",
			StartDate: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
			IsActive:  true,
		},
	}
	for _, y := range years {
		DB.Create(&y)
	}
}

func seedCategories() {
	type criteriaData struct {
		Name     string
		Desc     string
		MaxScore float64
	}

	type categoryData struct {
		Number   int
		Name     string
		Weight   float64
		IsBonus  bool
		Criteria []criteriaData
	}

	categories := []categoryData{
		{1, "اللغة", 50, false, []criteriaData{
			{"اللغة الإنكليزية", "ترجمة كاملة=25, ترجمة أبواب رئيسية=10, غير متوفرة=0", 25},
			{"جودة وسلامة لغة الموقع", "", 25},
		}},
		{2, "النطاق", 20, false, []criteriaData{
			{"النطاق الرسمي", "edu.iq=10, عادي=0", 10},
			{"البريد الالكتروني الرسمي", "", 10},
		}},
		{3, "احصائيات سنوية تراكمية", 40, false, []criteriaData{
			{"عدد الكليات", "", 10},
			{"عدد اعضاء هيئة التدريس", "", 10},
			{"عدد الطلبة", "", 10},
			{"عدد البحوث المنشورة", "", 10},
		}},
		{4, "تحديث الموقع", 30, false, []criteriaData{
			{"تحديث الاخبار", "", 10},
			{"خطة المؤتمرات والندوات والورش", "", 10},
			{"خطة الدورات التدريبية", "", 10},
		}},
		{5, "الوسائط المتعددة", 30, false, []criteriaData{
			{"فيديوات تعريفية", "", 10},
			{"ملفات صورية", "", 10},
			{"اصدارات إعلامية", "", 10},
		}},
		{6, "معلومات الاتصال", 40, false, []criteriaData{
			{"معلومات الاتصال بالجامعة", "", 10},
			{"معلومات الاتصال بالقسم", "", 10},
			{"استمارة اتصل بنا مع تاكيد الاستلام", "", 10},
			{"الربط بمواقع التواصل الاجتماعي", "", 10},
		}},
		{7, "روابط تعريفية عن الجامعة", 40, false, []criteriaData{
			{"نشأة الجامعة وتشكيلاتها", "", 10},
			{"تعريف عن الجامعة والهيكل التنظيمي", "", 10},
			{"رؤيا ورسالة واهداف الجامعة", "", 10},
			{"الأدلة والتعليمات والضوابط", "", 10},
		}},
		{8, "التنسيق والتصميم", 50, false, []criteriaData{
			{"تنسيق موحد للخط والحجم واللون", "", 10},
			{"سرعة التنقل والقوائم المنسدلة", "", 10},
			{"فاعلية الروابط", "", 10},
			{"خارطة الموقع site map", "", 20},
		}},
		{9, "مواقع ذات صلة", 40, false, []criteriaData{
			{"روابط مواقع الاقسام والكليات", "", 20},
			{"المواقع الصديقة", "", 20},
		}},
		{10, "الطلبة", 40, false, []criteriaData{
			{"وصف البرنامج الاكاديمي", "", 10},
			{"الطلبة الاوائل", "", 5},
			{"التقويم الجامعي", "", 5},
			{"التسجيل", "", 10},
			{"خطة القبول", "", 10},
		}},
		{11, "الخريجون", 50, false, []criteriaData{
			{"التواصل مع الخريجين", "", 10},
			{"التأهيل والتوظيف", "", 10},
			{"رأي أصحاب العمل", "", 10},
			{"الرموز العلمية والمشاهير", "", 10},
			{"الجوائز والأوسمة", "", 10},
		}},
		{12, "التدريسيون", 40, false, []criteriaData{
			{"دليل الكورسات", "", 10},
			{"اسماء التدريسيين", "", 10},
			{"السيرة العلمية", "", 10},
			{"وسيلة الاتصال بالتدريسي", "", 10},
		}},
		{13, "الجانب التقني", 60, false, []criteriaData{
			{"لغة برمجة الموقع", "Dynamic=20, Static=0", 20},
			{"صديق لمحركات البحث SEO", "", 10},
			{"ملائمة لجميع المتصفحات", "", 15},
			{"الموقع ملائم للأجهزة اللوحية والنقالة", "", 15},
		}},
		{14, "المنشورات", 30, false, []criteriaData{
			{"البحوث", "", 10},
			{"المجلات العلمية", "", 10},
			{"براءة الاختراع", "", 10},
		}},
		{15, "الخدمات الالكترونية", 100, false, []criteriaData{
			{"اسماء الخريجين واختصاصاتهم", "", 10},
			{"البريد الالكتروني الجامعي للطلاب", "", 10},
			{"الاستعلامات الالكترونية او ChatBot", "", 10},
			{"المكتبة الالكترونية", "", 20},
			{"بوابة الطالب Student Portal", "", 30},
			{"بوابة الأستاذ Staff Portal", "", 20},
		}},
		{16, "تصنيف الويبومتركس", 50, false, []criteriaData{
			{"التسلسل العالمي", "", 50},
		}},
		{17, "المنصات الالكترونية", 160, false, []criteriaData{
			{"نظام إدارة التعلم LMS", "", 20},
			{"نظام معلومات الطلبة SIS", "", 20},
			{"نظام ادارة الموارد ERP", "", 20},
			{"تطبيق موبايل", "", 20},
			{"منصة ادارة الترقيات العلمية", "", 20},
			{"منصة ادارة الأنشطة والفعاليات", "", 20},
			{"منصة ادارة المجلات العلمية", "", 20},
			{"بوابة الدفع الالكتروني", "", 20},
		}},
		{18, "المستودع الرقمي", 50, false, []criteriaData{
			{"مستودع رقمي للأصول الرقمية", "", 50},
		}},
		{19, "التعليم المستمر", 40, false, []criteriaData{
			{"منصة مساقات MOOC", "", 20},
			{"مساقات مفتوحة مع مؤسسات دولية", "", 20},
		}},
		{20, "الأمان والخصوصية", 40, false, []criteriaData{
			{"بروتوكول HTTPS وشهادة SSL", "", 10},
			{"جدار ناري Firewall", "", 10},
			{"إجراءات حماية البيانات", "", 10},
			{"حماية من الهجمات السيبرانية", "", 10},
		}},
		{21, "إضافي (BONUS)", 100, true, []criteriaData{
			{"بوابة للطلبة الأجانب", "", 20},
			{"أهداف التنمية المستدامة SDG", "", 20},
			{"اختبارات اختراق دورية", "", 20},
			{"حاضنة أعمال", "", 20},
			{"مبادرات إبداعية", "", 20},
		}},
	}

	for i, cat := range categories {
		category := models.Category{
			Number:    cat.Number,
			NameAr:    cat.Name,
			Weight:    cat.Weight,
			SortOrder: i + 1,
			IsBonus:   cat.IsBonus,
		}
		DB.Create(&category)

		for j, crit := range cat.Criteria {
			criteria := models.Criteria{
				CategoryID:  category.ID,
				NameAr:      crit.Name,
				Description: crit.Desc,
				MaxScore:    crit.MaxScore,
				SortOrder:   j + 1,
			}
			DB.Create(&criteria)
		}
	}
}

func seedUniversities() {
	govUniversities := []struct {
		Name    string
		Website string
	}{
		{"جامعة بغداد", "https://uobaghdad.edu.iq/"},
		{"الجامعة المستنصرية", "https://uomustansiriyah.edu.iq/"},
		{"الجامعة التكنولوجية", "https://uotechnology.edu.iq/"},
		{"جامعة النهرين", "https://nahrainuniv.edu.iq/"},
		{"جامعة الموصل", "https://uomosul.edu.iq/"},
		{"جامعة البصرة", "https://uobasrah.edu.iq/"},
		{"جامعة القادسية", "https://qu.edu.iq/"},
		{"جامعة الأنبار", "https://www.uoanbar.edu.iq/"},
		{"جامعة الكوفة", "https://uokufa.edu.iq/"},
		{"جامعة بابل", "https://uobabylon.edu.iq/"},
		{"جامعة ديالى", "https://uodiyala.edu.iq/"},
		{"جامعة كربلاء", "https://uokerbala.edu.iq/"},
		{"جامعة واسط", "https://uowasit.edu.iq/"},
		{"جامعة الفلوجة", "https://uofallujah.edu.iq/ar"},
		{"جامعة تكريت", "https://www.tu.edu.iq/index.php/ar/"},
		{"الجامعة العراقية", "https://aliraqia.edu.iq/"},
		{"جامعة ذي قار", "https://utq.edu.iq/"},
		{"جامعة كركوك", "https://uokirkuk.edu.iq/ar/"},
		{"جامعة ميسان", "https://uomisan.edu.iq/ar/"},
		{"جامعة المثنى", "https://mu.edu.iq/"},
		{"جامعة سامراء", "https://uosamarra.edu.iq/"},
		{"جامعة القاسم الخضراء", "https://uoqasim.edu.iq/"},
		{"جامعة سومر", "https://www.uos.edu.iq/"},
		{"جامعة نينوى", "https://uoninevah.edu.iq/"},
		{"جامعة الكرخ للعلوم", "https://kus.edu.iq/"},
		{"جامعة أبن سينا للعلوم الطبية والصيدلانية", "https://ibnsina.edu.iq/"},
		{"جامعة البصرة للنفط والغاز", "https://buog.edu.iq/ar/index.php/"},
		{"جامعة جابر بن حيان الطبية", "https://jmu.edu.iq/ar/"},
		{"جامعة الحمدانية", "https://www.uohamdaniya.edu.iq/"},
		{"جامعة تلعفر", "https://uotelafer.edu.iq/"},
		{"جامعة تكنولوجيا المعلومات والأتصالات", "https://uoitc.edu.iq/ar/"},
		{"الجامعة التقنية الشمالية", "https://ntu.edu.iq/"},
		{"الجامعة التقنية الجنوبية", "https://www.stu.edu.iq/"},
		{"الجامعة التقنية الوسطى", "https://atu.edu.iq/"},
		{"جامعة الفرات الأوسط التقنية", "https://atu.edu.iq/"},
		{"جامعة الشطرة", "https://shu.edu.iq/"},
	}

	for _, u := range govUniversities {
		uni := models.University{
			Name:    u.Name,
			Type:    "government",
			Website: u.Website,
		}
		DB.Create(&uni)
	}

	privateUniversities := []struct {
		Name    string
		Website string
	}{
		{"جامعة التراث", "https://uoturath.edu.iq"},
		{"كلية المنصور الجامعة", "https://muc.edu.iq"},
		{"جامعة الرافدين", "https://ruc.edu.iq"},
		{"جامعة المأمون", "https://almamonuc.edu.iq"},
		{"جامعة شط العرب", "https://sa-uc.edu.iq"},
		{"جامعة المعارف", "https://uoa.edu.iq"},
		{"جامعة الحدباء", "https://hu.edu.iq"},
		{"كلية بغداد للعلوم الاقتصادية", "https://baghdadcollege.edu.iq"},
		{"كلية اليرموك الجامعة", "https://al-yarmok.edu.iq"},
		{"كلية بغداد للعلوم الطبية", "https://bcms.edu.iq"},
		{"جامعة اهل البيت", "https://www.abu.edu.iq"},
		{"الجامعة الاسلامية", "https://iunajaf.edu.iq"},
		{"جامعة دجلة", "https://duc.edu.iq"},
		{"كلية السلام الجامعة", "https://alsalam.edu.iq"},
		{"جامعة الكفيل", "https://alkafeel.edu.iq"},
		{"جامعة مدينة العلم", "https://mauc.edu.iq"},
		{"جامعة الشيخ الطوسي", "https://altoosi.edu.iq"},
		{"جامعة الامام جعفر الصادق", "https://ijsu.edu.iq"},
		{"كلية الرشيد الجامعة", ""},
		{"كلية العراق الجامعة", "https://iraquniversity.net"},
		{"كلية صدر العراق الجامعة", "https://siuc.edu.iq"},
		{"جامعة القلم", "https://alqalam.edu.iq"},
		{"كلية الحسين الجامعة", "https://huciraq.edu.iq"},
		{"كلية الحكمة الجامعة", "https://hiuc.edu.iq"},
		{"جامعة المستقبل", "https://uomus.edu.iq"},
		{"كلية الحضارة الجامعة", "https://alimamunc.edu.iq"},
		{"جامعة الحلة", "https://hilla-unc.edu.iq"},
		{"كلية اصول العلم الجامعة", "https://ouc.edu.iq"},
		{"جامعة الاسراء", "https://esraa.edu.iq"},
		{"جامعة الصفوة", "https://alsafwa.edu.iq"},
		{"جامعة الكتاب", "https://uoalkitab.edu.iq"},
		{"جامعة الكوت", "https://alkutcollege.edu.iq"},
		{"جامعة الفراهيدي", "https://uoalfarahidi.edu.iq"},
		{"جامعة المصطفى", "https://almustafauniversity.edu.iq"},
		{"كلية مزايا الجامعة", "https://mpu.edu.iq"},
		{"جامعة النور", "https://alnoor.edu.iq"},
		{"جامعة الكنوز", "https://kunoozu.edu.iq"},
		{"جامعة الفارابي", "https://alfarabiuc.edu.iq"},
		{"كلية الباني الجامعة", "https://albani.edu.iq"},
		{"كلية الطف", "https://altuff.edu.iq"},
		{"جامعة الزهراوي", "https://alzahu.edu.iq"},
		{"كلية النخبة الجامعة", "https://alnukhba.edu.iq"},
		{"جامعة النسور", "https://nuc.edu.iq"},
		{"جامعة بلاد الرافدين", "https://www.bauc14.edu.iq"},
		{"جامعة الفرقدين", "https://fu.edu.iq"},
		{"جامعة اوروك", "https://uruk.edu.iq"},
		{"جامعة الهادي", "https://huc.edu.iq"},
		{"جامعة البيان", "https://albayan.edu.iq"},
		{"جامعة وارث الانبياء", "https://uowa.edu.iq"},
		{"جامعة الامين", "https://alameen.edu.iq"},
		{"جامعة العميد", "https://alameed.edu.iq"},
		{"جامعة اشور", "https://au.edu.iq"},
		{"جامعة المنارة", "https://uomanara.edu.iq"},
		{"جامعة العين العراقية", "https://alayen.edu.iq"},
		{"كلية الشرق الاوسط", "https://Meuc.edu.iq"},
		{"كلية العمارة", "https://alamarhuc.edu.iq"},
		{"جامعة الزهراء للبنات", "https://alzahraa.edu.iq"},
		{"جامعة كلكامش", "https://gu.edu.iq"},
		{"الجامعة الامريكية", "https://auib.edu.iq"},
		{"جامعة المعقل", "https://www.almaaqal.edu.iq"},
		{"جامعة المشرق", "https://uom.edu.iq"},
		{"كلية ابن خلدون", "https://ik.edu.iq"},
		{"كلية الهدى", "https://uoalhuda.edu.iq"},
		{"جامعة ساوة", "https://www.sawa-un.edu.iq"},
		{"جامعة الشعب", "https://alshaab.edu.iq"},
		{"جامعة النبراس", "https://nibru.edu.iq"},
		{"جامعة قرطبة", "https://cur.edu.iq"},
		{"جامعة بابا كركر", "https://babagurguruni.edu.iq"},
		{"جامعة الفرقان", "https://alfurqan.edu.iq"},
	}

	for _, u := range privateUniversities {
		uni := models.University{
			Name:    u.Name,
			Type:    "private",
			Website: u.Website,
		}
		DB.Create(&uni)
	}
}

func seedUniversityUsers() {
	defaultPassword := "Mohe@2025"
	hash, _ := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)

	var universities []models.University
	DB.Find(&universities)

	created := 0
	usedUsernames := make(map[string]bool)

	for _, uni := range universities {
		// Extract domain from website URL for username and email
		domain := ""
		if uni.Website != "" {
			parsed, err := url.Parse(uni.Website)
			if err == nil && parsed.Host != "" {
				domain = parsed.Host
				domain = strings.TrimPrefix(domain, "www.")
			}
		}

		// Username = domain (e.g. uoturath.edu.iq)
		username := domain
		if username == "" {
			username = fmt.Sprintf("uni_%d", uni.ID)
		}

		// Handle duplicate domains
		if usedUsernames[username] {
			username = fmt.Sprintf("%s_%d", username, uni.ID)
		}
		usedUsernames[username] = true

		email := fmt.Sprintf("info@%s", domain)
		if domain == "" {
			email = fmt.Sprintf("uni_%d@university.edu.iq", uni.ID)
		}

		user := models.User{
			Username:           username,
			Password:           string(hash),
			FullName:           uni.Name,
			Email:              email,
			Phone:              "0000",
			Role:               "university",
			MustChangePassword: true,
			UniversityID:       &uni.ID,
		}

		// Update university contact info
		DB.Model(&uni).Updates(map[string]interface{}{
			"contact_email": email,
			"contact_phone": "0000",
		})

		DB.Create(&user)
		created++
	}

	log.Printf("Created %d university user accounts (username=domain, default password: %s)", created, defaultPassword)
}
