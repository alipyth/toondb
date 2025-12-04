# TOON DB 🚀

## English Version

A high-performance, lightweight Key-Value database for the AI era.

[![Go Version](https://img.shields.io/badge/Go-1.21-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://www.docker.com/)

TOON DB is a modern Key-Value database specifically designed for storing and retrieving data in TOON format. This format maintains human readability while significantly reducing token consumption when interacting with Language Models (LLMs).

### ✨ Key Features

- 🚀 **Ultra-fast Performance**: Core written in Go using the powerful BadgerDB engine.
- 🛡️ **Layered Security**: Mandatory API Key authentication for all requests and admin panel.
- 📝 **TOON Format**: Built-in parser for automatic conversion between TOON and JSON formats.
- 🖥️ **Visual Management Panel**: Web interface for viewing, editing, deleting, and managing backups.
- 💾 **Backup & Restore**: One-click database export and import functionality.
- 🔄 **Atomic Operations**: Support for secure data storage transactions.

### 🚀 Quick Start

#### Prerequisites
- Docker and Docker Compose

#### 1. Run with Docker
Simply run the following command in the project root:

```bash
docker-compose up -d --build
```

The service will be available at http://localhost:3000.

#### 2. Security Configuration (Important)
In the docker-compose.yml file, change the API_KEY value:

```yaml
environment:
  - API_KEY=your-super-secret-key
```

Default key: `toondb-secure-key`

### 🖥 Management Panel Guide

1. Open your browser and go to http://localhost:3000.
2. Enter your API Key on the login page.
3. After successful login, you can:
   - View collections and keys
   - Edit and save data (Update)
   - Create new data (Create)
   - Delete entire collections
   - View all keys in a collection
   - Take database backups or restore backup files

### 📚 API Documentation (with Examples)

All requests must include the `X-API-Key` header.

#### 1. Check Status and Authentication
```bash
curl -H "X-API-Key: toondb-secure-key" http://localhost:3000/api/auth
```

#### 2. Insert or Update Data (Upsert)
To create new data or edit existing data, use the POST method. The request body must be in TOON format.

Example: Save user information in the users collection with key ali:

```bash
curl -X POST http://localhost:3000/api/users/ali \
  -H "X-API-Key: toondb-secure-key" \
  -H "Content-Type: text/plain" \
  -d "name: Ali Rezaei\nage: 28\nskills[3]: go,python,docker\ncontact:\n  email: ali@example.com\n  phone: +989120000000"
```

Note: If the ali key already exists, the new data will replace it (Update).

#### 3. Read Data
Retrieve data in TOON format:

```bash
curl -H "X-API-Key: toondb-secure-key" http://localhost:3000/api/users/ali
```

#### 4. Delete Data
```bash
curl -X DELETE http://localhost:3000/api/users/ali \
  -H "X-API-Key: toondb-secure-key"
```

#### 5. List All Data
View all collections and keys:

```bash
curl -H "X-API-Key: toondb-secure-key" http://localhost:3000/api/collections
```

#### 6. Get All Keys in a Collection
```bash
curl -H "X-API-Key: toondb-secure-key" http://localhost:3000/api/collections/users
```

#### 7. Delete Entire Collection
```bash
curl -X DELETE http://localhost:3000/api/collections/users \
  -H "X-API-Key: toondb-secure-key"
```

### 💻 Code Examples (Python & Node.js)

#### Python (Simple Script)
```python
import requests

API_URL = "http://localhost:3000/api"
API_KEY = "toondb-secure-key"
HEADERS = {"X-API-Key": API_KEY, "Content-Type": "text/plain"}

# 1. Save or update data (Upsert)
toon_data = """
name: Sara
role: Data Scientist
skills[2]: python,pytorch
"""
response = requests.post(f"{API_URL}/users/sara", data=toon_data, headers=HEADERS)
print("Save Status:", response.json())

# 2. Read data
response = requests.get(f"{API_URL}/users/sara", headers=HEADERS)
print("\nData Received:\n", response.text)
```

#### Python (Wrapper Class)
For cleaner and more convenient use in large projects:

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

    def get_collection_keys(self, collection):
        """Get all keys in a collection"""
        url = f"{self.base_url}/api/collections/{collection}"
        resp = requests.get(url, headers=self.headers)
        return resp.json().get('data', {}).get('keys', []) if resp.status_code == 200 else []

    def delete_collection(self, collection):
        """Delete an entire collection"""
        url = f"{self.base_url}/api/collections/{collection}"
        resp = requests.delete(url, headers=self.headers)
        return resp.status_code == 200

# Usage:
db = ToonDB()

# Save
db.save("myresume", "personal", "name: Ali\njob: Developer")

# Read
print(db.get("myresume", "personal"))

# Get all keys in collection
print(db.get_collection_keys("myresume"))

# Delete collection
# db.delete_collection("myresume")
```

#### Node.js
```javascript
const API_URL = 'http://localhost:3000/api';
const API_KEY = 'toondb-secure-key';

async function main() {
  const headers = { 'X-API-Key': API_KEY, 'Content-Type': 'text/plain' };

  // 1. Save or update data
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

  // 2. Read data
  const readRes = await fetch(`${API_URL}/users/mike`, { headers });
  console.log('\nData Received:\n', await readRes.text());
}

main();
```

### 💾 Backup and Restore Management

#### Create Backup
This command downloads a JSON file containing all database data:

```bash
curl -H "X-API-Key: toondb-secure-key" http://localhost:3000/api/backup > backup.json
```

#### Restore Data
Warning: This operation will replace/merge existing data with backup file data.

```bash
curl -X POST http://localhost:3000/api/restore \
  -H "X-API-Key: toondb-secure-key" \
  -H "Content-Type: application/json" \
  -d @backup.json
```

### 📝 Introduction to TOON Format

The TOON format is similar to YAML but simpler:

```toon
# Simple key-value
title: Project Manager

# Simple array
tags[3]: urgent,backend,api

# Nested object
metadata:
  created: 2023-10-01
  author: admin

# Array of objects (table)
users[2]{id,name}:
  1,Ali
  2,Sara
```

---

## نسخه فارسی

تون دی‌بی (TOON DB) یک دیتابیس Key-Value مدرن است که به طور اختصاصی برای ذخیره‌سازی و بازیابی داده‌ها با فرمت TOON طراحی شده است. این فرمت با ساختار فشرده خود، خوانایی را برای انسان حفظ کرده و مصرف توکن‌ها را در تعامل با مدل‌های زبانی (LLM) به شدت کاهش می‌دهد.

### ✨ ویژگی‌های کلیدی

- 🚀 **عملکرد فوق‌سریع**: هسته نوشته شده با Go و استفاده از موتور قدرتمند BadgerDB.
- 🛡️ **امنیت لایه‌ای**: احراز هویت اجباری با API Key برای تمام درخواست‌ها و پنل مدیریت.
- 📝 **فرمت TOON**: پارسر داخلی برای تبدیل خودکار فرمت TOON به JSON و برعکس.
- 🖥️ **پنل مدیریت بصری**: رابط کاربری وب برای مشاهده، ویرایش، حذف و مدیریت بکاپ‌ها.
- 💾 **بکاپ و ریستور**: قابلیت خروجی گرفتن از کل دیتابیس و بازگردانی آن با یک کلیک.
- 🔄 **عملیات اتمیک**: پشتیبانی از تراکنش‌های امن برای ذخیره‌سازی داده‌ها.

### 🚀 نصب و اجرا (Quick Start)

#### پیش‌نیازها
- Docker و Docker Compose

#### ۱. اجرا با داکر
کافیست دستور زیر را در ریشه پروژه اجرا کنید:

```bash
docker-compose up -d --build
```

سرویس در آدرس http://localhost:3000 در دسترس خواهد بود.

#### ۲. تنظیمات امنیتی (مهم)
در فایل docker-compose.yml، مقدار API_KEY را تغییر دهید:

```yaml
environment:
  - API_KEY=your-super-secret-key
```

کلید پیش‌فرض: `toondb-secure-key`

### 🖥 راهنمای پنل مدیریت

۱. مرورگر را باز کنید و به http://localhost:3000 بروید.
۲. در صفحه ورود، API Key خود را وارد کنید.
۳. پس از ورود موفق، می‌توانید:
   - کالکشن‌ها و کلیدها را مشاهده کنید.
   - داده‌ها را ویرایش و ذخیره کنید (Update).
   - داده‌های جدید بسازید (Create).
   - کل کالکشن را حذف کنید.
   - تمام کلیدهای یک کالکشن را مشاهده کنید.
   - از دیتابیس بکاپ بگیرید یا فایل بکاپ را ریستور کنید.

### 📚 مستندات API (با مثال)

تمام درخواست‌ها باید دارای هدر `X-API-Key` باشند.

#### ۱. بررسی وضعیت و احراز هویت
```bash
curl -H "X-API-Key: toondb-secure-key" http://localhost:3000/api/auth
```

#### ۲. ثبت یا آپدیت داده (Upsert)
برای ساخت داده جدید یا ویرایش داده موجود، از متد POST استفاده کنید. بدنه درخواست باید متن با فرمت TOON باشد.

مثال: ذخیره اطلاعات یک کاربر در کالکشن users با کلید ali:

```bash
curl -X POST http://localhost:3000/api/users/ali \
  -H "X-API-Key: toondb-secure-key" \
  -H "Content-Type: text/plain" \
  -d "name: Ali Rezaei\nage: 28\nskills[3]: go,python,docker\ncontact:\n  email: ali@example.com\n  phone: +989120000000"
```

نکته: اگر کلید ali از قبل وجود داشته باشد، داده‌های جدید جایگزین می‌شوند (Update).

#### ۳. خواندن داده (Read)
دریافت داده به فرمت TOON:

```bash
curl -H "X-API-Key: toondb-secure-key" http://localhost:3000/api/users/ali
```

#### ۴. حذف داده (Delete)
```bash
curl -X DELETE http://localhost:3000/api/users/ali \
  -H "X-API-Key: toondb-secure-key"
```

#### ۵. لیست کردن تمام داده‌ها
مشاهده تمام کالکشن‌ها و کلیدها:

```bash
curl -H "X-API-Key: toondb-secure-key" http://localhost:3000/api/collections
```

#### ۶. دریافت تمام کلیدهای یک کالکشن
```bash
curl -H "X-API-Key: toondb-secure-key" http://localhost:3000/api/collections/users
```

#### ۷. حذف کل کالکشن
```bash
curl -X DELETE http://localhost:3000/api/collections/users \
  -H "X-API-Key: toondb-secure-key"
```

### 💻 نمونه کدها (Python & Node.js)

#### Python (اسکریپت ساده)
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

#### Python (کلاس Wrapper)
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

    def get_collection_keys(self, collection):
        """Get all keys in a collection"""
        url = f"{self.base_url}/api/collections/{collection}"
        resp = requests.get(url, headers=self.headers)
        return resp.json().get('data', {}).get('keys', []) if resp.status_code == 200 else []

    def delete_collection(self, collection):
        """Delete an entire collection"""
        url = f"{self.base_url}/api/collections/{collection}"
        resp = requests.delete(url, headers=self.headers)
        return resp.status_code == 200

# استفاده:
db = ToonDB()

# ذخیره
db.save("myresume", "personal", "name: Ali\njob: Developer")

# خواندن
print(db.get("myresume", "personal"))

# دریافت تمام کلیدهای کالکشن
print(db.get_collection_keys("myresume"))

# حذف کالکشن
# db.delete_collection("myresume")
```

#### Node.js
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

### 💾 مدیریت بکاپ و ریستور

#### تهیه نسخه پشتیبان (Backup)
این دستور یک فایل JSON شامل تمام داده‌های دیتابیس را دانلود می‌کند:

```bash
curl -H "X-API-Key: toondb-secure-key" http://localhost:3000/api/backup > backup.json
```

#### بازگردانی اطلاعات (Restore)
هشدار: این عملیات داده‌های موجود را با داده‌های فایل بکاپ جایگزین/ادغام می‌کند.

```bash
curl -X POST http://localhost:3000/api/restore \
  -H "X-API-Key: toondb-secure-key" \
  -H "Content-Type: application/json" \
  -d @backup.json
```

### 📝 آشنایی با فرمت TOON

فرمت TOON شبیه به YAML اما ساده‌تر است:

```toon
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

## 🏗️ Project Structure

```
toon-db/
├── cmd/server/main.go          # Main application entry point
├── internal/
│   ├── db/database.go          # Database layer with BadgerDB
│   ├── parser/toon.go          # TOON format parser
│   └── handlers/handlers.go    # API and web handlers
├── web/                        # Static web files
├── Dockerfile                  # Docker configuration
├── docker-compose.yml          # Docker Compose configuration
├── go.mod                      # Go module
└── README.md                   # Documentation
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit an Issue or a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📞 Contact

- **Developer**: Ali Jahani
- **Website**: https://jahaniwww.com
- **Email**: [satreyek@gmail.com](mailto:satreyek@gmail.com)

---

⭐ If you find this project useful, please give it a star!
