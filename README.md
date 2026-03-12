# نظام تقييم جودة المواقع الالكترونية الجامعية
## University Website Quality Evaluation System

<div align="center">

![Ministry of Higher Education](frontend/public/mohesr-logo.svg)

**وزارة التعليم العالي والبحث العلمي - جمهورية العراق**

نظام متكامل لتقييم جودة المواقع الالكترونية للجامعات العراقية الحكومية والأهلية

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev)
[![Vue.js](https://img.shields.io/badge/Vue.js-3-4FC08D?style=for-the-badge&logo=vue.js&logoColor=white)](https://vuejs.org)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://docker.com)
[![Coolify](https://img.shields.io/badge/Coolify-Deploy-6C3FC5?style=for-the-badge)](https://coolify.io)

</div>

---

## المميزات

- **نموذج تقييم شامل** - 20 فئة تقييم + فئة إضافية (1000 + 100 درجة) مطابق لاستمارة الوزارة
- **إعادة التقديم** مع مقارنة الفروقات بين النسخ (Diff)
- **نظام صلاحيات متعدد المستويات** - سوبر أدمن، مراجع (أدمن ثانوي)، جامعة
- **تخصيص المراجعين** - تحديد فئات معينة لكل مراجع
- **سير عمل المراجعة** - مسودة ← مقدم ← قيد المراجعة ← معتمد/مرفوض
- **إحصائيات ورسوم بيانية** - مقارنة الجامعات، تصنيفات، متوسطات
- **ذكاء اصطناعي** - تحليل واقتراحات عبر DeepSeek و Gemini
- **تصنيف حسب السنة الدراسية**
- **واجهة عربية كاملة** مع دعم RTL
- **بيانات مسبقة** - جميع الجامعات الحكومية (36) والأهلية (68+) مُدخلة

## التقنيات المستخدمة

| الطبقة | التقنية |
|--------|---------|
| Backend | Go, Gin, GORM, SQLite, JWT |
| Frontend | Vue 3, Vuetify 4, Pinia, Chart.js |
| Deployment | Docker, Docker Compose, Nginx |
| AI | DeepSeek API, Google Gemini API |

## التشغيل السريع

### باستخدام Docker Compose (موصى به)

```bash
# 1. استنسخ المشروع
git clone https://github.com/haydary1986/university-website-eval.git
cd university-website-eval

# 2. أنشئ ملف البيئة
cp .env.example .env
# عدّل .env وأضف JWT_SECRET

# 3. شغّل
docker compose up -d

# التطبيق يعمل على http://localhost:3000
```

### التشغيل المحلي للتطوير

```bash
# Backend
cd backend
go mod tidy
go run .
# يعمل على http://localhost:8080

# Frontend (في terminal ثاني)
cd frontend
npm install
npm run dev
# يعمل على http://localhost:5173
```

## النشر على Coolify

1. ارفع المشروع على GitHub
2. في Coolify: **New Resource → Docker Compose**
3. اربط الـ Repository
4. أضف المتغيرات البيئية:

| المتغير | مطلوب | الوصف |
|---------|-------|-------|
| `JWT_SECRET` | نعم | مفتاح سري لتوقيع JWT (ولّده بـ `openssl rand -base64 32`) |
| `DEEPSEEK_API_KEY` | لا | مفتاح DeepSeek API للذكاء الاصطناعي |
| `GEMINI_API_KEY` | لا | مفتاح Google Gemini API للذكاء الاصطناعي |
| `APP_PORT` | لا | منفذ التطبيق (الافتراضي: 3000) |

5. انشر!

## بيانات الدخول الافتراضية

| الحقل | القيمة |
|-------|--------|
| اسم المستخدم | `admin` |
| كلمة المرور | `Admin@2024` |
| الدور | مدير عام (Super Admin) |

> **تنبيه:** غيّر كلمة المرور فوراً بعد أول تسجيل دخول في بيئة الإنتاج

## هيكل المشروع

```
├── docker-compose.yml          # إعدادات Docker
├── .env.example                # نموذج متغيرات البيئة
├── coolify.json                # إعدادات Coolify
│
├── backend/                    # Go API
│   ├── main.go                 # نقطة الدخول
│   ├── config/                 # إعدادات النظام
│   ├── models/                 # نماذج قاعدة البيانات
│   ├── database/               # اتصال DB + بيانات أولية
│   ├── middleware/             # JWT + صلاحيات
│   ├── handlers/               # معالجات API
│   └── services/               # خدمة الذكاء الاصطناعي
│
└── frontend/                   # Vue 3 SPA
    ├── nginx.conf              # إعدادات Nginx
    └── src/
        ├── views/              # الصفحات (11 صفحة)
        ├── components/         # مكونات مشتركة
        ├── stores/             # إدارة الحالة (Pinia)
        ├── services/           # عميل API
        └── router/             # التوجيه + حراسة المسارات
```

## فئات التقييم (20 فئة + إضافية)

| # | الفئة | الوزن |
|---|-------|-------|
| 1 | اللغة | 50 |
| 2 | النطاق | 20 |
| 3 | إحصائيات سنوية تراكمية | 40 |
| 4 | تحديث الموقع | 30 |
| 5 | الوسائط المتعددة | 30 |
| 6 | معلومات الاتصال | 40 |
| 7 | روابط تعريفية عن الجامعة | 40 |
| 8 | التنسيق والتصميم | 50 |
| 9 | مواقع ذات صلة | 40 |
| 10 | الطلبة | 40 |
| 11 | الخريجون | 50 |
| 12 | التدريسيون | 40 |
| 13 | الجانب التقني | 60 |
| 14 | المنشورات | 30 |
| 15 | الخدمات الالكترونية | 100 |
| 16 | تصنيف الويبومتركس | 50 |
| 17 | المنصات الالكترونية | 160 |
| 18 | المستودع الرقمي | 50 |
| 19 | التعليم المستمر | 40 |
| 20 | الأمان والخصوصية | 40 |
| **إضافية** | **نقاط قوة إضافية** | **100** |
| | **المجموع** | **1000 + 100** |

## API Endpoints

<details>
<summary>عرض جميع المسارات</summary>

### المصادقة
- `POST /api/auth/login` - تسجيل الدخول

### الجامعات
- `GET /api/universities` - قائمة الجامعات
- `GET /api/universities/:id` - تفاصيل جامعة
- `PUT /api/universities/:id` - تحديث جامعة

### السنوات الدراسية
- `GET /api/academic-years` - القائمة
- `POST /api/academic-years` - إضافة (سوبر أدمن)
- `PUT /api/academic-years/:id` - تعديل (سوبر أدمن)

### فئات التقييم
- `GET /api/categories` - القائمة مع المعايير
- `GET /api/categories/:id` - تفاصيل فئة

### التقديمات
- `GET /api/submissions` - القائمة
- `POST /api/submissions` - إنشاء (جامعة)
- `PUT /api/submissions/:id` - تعديل مسودة (جامعة)
- `POST /api/submissions/:id/submit` - تقديم للمراجعة
- `GET /api/submissions/:id/diff/:version` - مقارنة النسخ

### المراجعة (أدمن)
- `GET /api/admin/submissions` - تقديمات للمراجعة
- `POST /api/admin/submissions/:id/review` - إرسال الدرجات
- `PUT /api/admin/submissions/:id/approve` - اعتماد
- `PUT /api/admin/submissions/:id/reject` - رفض

### إدارة المستخدمين (سوبر أدمن)
- `GET /api/admin/users` - القائمة
- `POST /api/admin/users` - إضافة
- `PUT /api/admin/users/:id` - تعديل
- `DELETE /api/admin/users/:id` - حذف
- `PUT /api/admin/users/:id/assign-categories` - تخصيص فئات

### الإحصائيات
- `GET /api/stats/overview` - نظرة عامة
- `GET /api/stats/universities` - تصنيف الجامعات
- `GET /api/stats/categories` - متوسطات الفئات

### الذكاء الاصطناعي
- `POST /api/ai/analyze-submission/:id` - تحليل تقديم
- `POST /api/ai/suggest-improvements/:id` - اقتراح تحسينات
- `POST /api/ai/compare-universities` - مقارنة جامعات

</details>

## الترخيص

هذا المشروع مطور لصالح وزارة التعليم العالي والبحث العلمي - جمهورية العراق.
