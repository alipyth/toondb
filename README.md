# TOON DB 🚀

**دیتابیس فوق‌سریع و کم‌حجم برای عصر هوش مصنوعی**

[![Go Version](https://img.shields.io/badge/Go-1.21-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://www.docker.com/)

**تون دی‌بی (TOON DB)** یک دیتابیس Key-Value مدرن است که به طور اختصاصی برای ذخیره‌سازی و بازیابی داده‌ها با فرمت **TOON** طراحی شده است. این فرمت با ساختار فشرده خود، خوانایی را برای انسان حفظ کرده و مصرف توکن‌ها را در تعامل با مدل‌های زبانی (LLM) به شدت کاهش می‌دهد.

---

## ✨ ویژگی‌های کلیدی

*   **🚀 عملکرد فوق‌سریع:** هسته نوشته شده با Go و استفاده از موتور قدرتمند BadgerDB.
*   **🛡️ امنیت لایه‌ای:** احراز هویت اجباری با API Key برای تمام درخواست‌ها و پنل مدیریت.
*   **📝 فرمت TOON:** پارسر داخلی برای تبدیل خودکار فرمت TOON به JSON و برعکس.
*   **🖥️ پنل مدیریت بصری:** رابط کاربری وب برای مشاهده، ویرایش، حذف و مدیریت بکاپ‌ها.
*   **💾 بکاپ و ریستور:** قابلیت خروجی گرفتن از کل دیتابیس و بازگردانی آن با یک کلیک.
*   **🔄 عملیات اتمیک:** پشتیبانی از تراکنش‌های امن برای ذخیره‌سازی داده‌ها.

---

## 🚀 نصب و اجرا (Quick Start)

### پیش‌نیازها
*   Docker و Docker Compose

### ۱. اجرا با داکر
کافیست دستور زیر را در ریشه پروژه اجرا کنید:
```bash
docker-compose up -d --build
```
سرویس در آدرس `http://localhost:3000` در دسترس خواهد بود.

### ۲. تنظیمات امنیتی (مهم)
در فایل `docker-compose.yml`، مقدار `API_KEY` را تغییر دهید:
```yaml
environment:
  - API_KEY=your-super-secret-key
```
*کلید پیش‌فرض: `toondb-secure-key`*

---

## 🖥 راهنمای پنل مدیریت

۱. مرورگر را باز کنید و به `http://localhost:3000` بروید.
۲. در صفحه ورود، **API Key** خود را وارد کنید.
۳. پس از ورود موفق، می‌توانید:
    *   کالکشن‌ها و کلیدها را مشاهده کنید.
    *   داده‌ها را ویرایش و ذخیره کنید (Update).
    *   داده‌های جدید بسازید (Create).
    *   از دیتابیس بکاپ بگیرید یا فایل بکاپ را ریستور کنید.

---

## 📚 مستندات API (با مثال)

تمام درخواست‌ها باید دارای هدر `X-API-Key` باشند.

### ۱. بررسی وضعیت و احراز هویت
```bash
curl -H "X-API-Key: toondb-secure-key" http://localhost:3000/api/auth
```

### ۲. ثبت یا آپدیت داده (Upsert)
برای ساخت داده جدید یا ویرایش داده موجود، از متد `POST` استفاده کنید. بدنه درخواست باید متن با فرمت **TOON** باشد.

**مثال:** ذخیره اطلاعات یک کاربر در کالکشن `users` با کلید `ali`:
```bash
curl -X POST http://localhost:3000/api/users/ali \
  -H "X-API-Key: toondb-secure-key" \
  -H "Content-Type: text/plain" \
  -d "name: Ali Rezaei\nage: 28\nskills[3]: go,python,docker\ncontact:\n  email: ali@example.com\n  phone: +989120000000"
```
> **نکته:** اگر کلید `ali` از قبل وجود داشته باشد، داده‌های جدید جایگزین می‌شوند (Update).

### ۳. خواندن داده (Read)
دریافت داده به فرمت TOON:
```bash
curl -H "X-API-Key: toondb-secure-key" http://localhost:3000/api/users/ali
```

### ۴. حذف داده (Delete)
```bash
curl -X DELETE http://localhost:3000/api/users/ali \
  -H "X-API-Key: toondb-secure-key"
```

### ۵. لیست کردن تمام داده‌ها
مشاهده تمام کالکشن‌ها و کلیدها:
```bash
curl -H "X-API-Key: toondb-secure-key" http://localhost:3000/api/collections
```

---

## 💻 نمونه کدها (Python & Node.js)

### Python (اسکریپت ساده)
```python
import requests

API_URL = "http://localhost:3000/api"
API_KEY = "toondb-secure-key"
HEADERS = {"X-API-Key": API_KEY, "Content-Type": "text/plain"}

# 1. ذخیره یا آپدیت داده (Upsert)
toon_data = """
name: Sara
role: Data Scientist
skills[2]: python,pytorch
"""
response = requests.post(f"{API_URL}/users/sara", data=toon_data, headers=HEADERS)
print("Save Status:", response.json())

# 2. خواندن داده
response = requests.get(f"{API_URL}/users/sara", headers=HEADERS)
print("\nData Received:\n", response.text)
```

### Python (کلاس Wrapper)
برای استفاده راحت‌تر و تمیزتر در پروژه‌های بزرگ:
```python
import requests

class ToonDB:
    def __init__(self, base_url="http://localhost:3000", api_key="toondb-secure-key"):
        self.base_url = base_url
        self.headers = {"X-API-Key": api_key, "Content-Type": "text/plain"}

    def save(self, collection, key, toon_data):
        """Create or Update a record"""
        url = f"{self.base_url}/api/{collection}/{key}"
        resp = requests.post(url, data=toon_data, headers=self.headers)
        return resp.status_code == 200

    def get(self, collection, key):
        """Retrieve a record in TOON format"""
        url = f"{self.base_url}/api/{collection}/{key}"
        resp = requests.get(url, headers=self.headers)
        return resp.text if resp.status_code == 200 else None

    def delete(self, collection, key):
        """Delete a record"""
        url = f"{self.base_url}/api/{collection}/{key}"
        resp = requests.delete(url, headers=self.headers)
        return resp.status_code == 200

# استفاده:
db = ToonDB()

# ذخیره
db.save("myresume", "personal", "name: Ali\njob: Developer")

# خواندن
print(db.get("myresume", "personal"))

# حذف
# db.delete("myresume", "personal")
```

### Node.js
```javascript
const API_URL = 'http://localhost:3000/api';
const API_KEY = 'toondb-secure-key';

async function main() {
  const headers = { 'X-API-Key': API_KEY, 'Content-Type': 'text/plain' };

  // 1. ذخیره یا آپدیت داده
  const toonData = `
name: Mike
role: Frontend Dev
skills[2]: react,css
  `;
  
  const saveRes = await fetch(`${API_URL}/users/mike`, {
    method: 'POST',
    headers: headers,
    body: toonData
  });
  console.log('Save Status:', await saveRes.json());

  // 2. خواندن داده
  const readRes = await fetch(`${API_URL}/users/mike`, { headers });
  console.log('\nData Received:\n', await readRes.text());
}

main();
```

---

## 💾 مدیریت بکاپ و ریستور

### تهیه نسخه پشتیبان (Backup)
این دستور یک فایل JSON شامل تمام داده‌های دیتابیس را دانلود می‌کند:
```bash
curl -H "X-API-Key: toondb-secure-key" http://localhost:3000/api/backup > backup.json
```

### بازگردانی اطلاعات (Restore)
**هشدار:** این عملیات داده‌های موجود را با داده‌های فایل بکاپ جایگزین/ادغام می‌کند.
```bash
curl -X POST http://localhost:3000/api/restore \
  -H "X-API-Key: toondb-secure-key" \
  -H "Content-Type: application/json" \
  -d @backup.json
```

---

## 📝 آشنایی با فرمت TOON
فرمت TOON شبیه به YAML اما ساده‌تر است:

```text
# کلید و مقدار ساده
title: Project Manager

# آرایه ساده
tags[3]: urgent,backend,api

# آبجکت تو در تو
metadata:
  created: 2023-10-01
  author: admin

# آرایه ای از آبجکت ها (جدول)
users[2]{id,name}:
  1,Ali
  2,Sara
```
