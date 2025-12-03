# TOON DB 🚀

**High-Performance Key-Value Store Optimized for AI & TOON Format**

[![Go Version](https://img.shields.io/badge/Go-1.21-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://www.docker.com/)

**TOON DB** is a modern Key-Value database designed to store and retrieve data in the **TOON** format. This format is human-readable, compact, and optimized to reduce token usage when interacting with Large Language Models (LLMs).

---

## ✨ Key Features

*   **🚀 Blazing Fast:** Built with Go and powered by the BadgerDB engine.
*   **🛡️ Secure:** Mandatory API Key authentication for all endpoints and the dashboard.
*   **📝 TOON Native:** Built-in parser to convert TOON format to JSON and vice-versa.
*   **🖥️ Web Dashboard:** Visual interface to view, edit, delete, and manage backups.
*   **💾 Backup & Restore:** Full database dump and restore capabilities via API and UI.
*   **🔄 Atomic Operations:** Safe data storage with transaction support.

---

## 🚀 Quick Start

### Prerequisites
*   Docker & Docker Compose

### 1. Run with Docker
Execute the following command in the project root:
```bash
docker-compose up -d --build
```
The service will be available at `http://localhost:3000`.

### 2. Security Configuration
Update the `API_KEY` in your `docker-compose.yml` file:
```yaml
environment:
  - API_KEY=your-super-secret-key
```
*Default key: `toondb-secure-key`*

---

## 🖥 Dashboard Guide

1. Open your browser and go to `http://localhost:3000`.
2. Enter your **API Key** in the login screen.
3. Once authenticated, you can:
    *   Browse collections and keys.
    *   Edit and Save data (Update).
    *   Create new entries.
    *   Download backups or restore from a file.

---

## 📚 API Reference

All requests must include the `X-API-Key` header.

### 1. Authentication Check
```bash
curl -H "X-API-Key: toondb-secure-key" http://localhost:3000/api/auth
```

### 2. Create or Update Data (Upsert)
Use the `POST` method to create a new key or update an existing one. The body should be in **TOON** format.

**Example:** Save user data in collection `users` with key `john`:
```bash
curl -X POST http://localhost:3000/api/users/john \
  -H "X-API-Key: toondb-secure-key" \
  -H "Content-Type: text/plain" \
  -d "name: John Doe\nrole: developer\nskills[2]: rust,go\ncontact:\n  email: john@example.com"
```
> **Note:** If the key `john` already exists, the data will be overwritten (Updated).

### 3. Read Data
Retrieve data in TOON format:
```bash
curl -H "X-API-Key: toondb-secure-key" http://localhost:3000/api/users/john
```

### 4. Delete Data
```bash
curl -X DELETE http://localhost:3000/api/users/john \
  -H "X-API-Key: toondb-secure-key"
```

### 5. List All Data
View all collections and keys:
```bash
curl -H "X-API-Key: toondb-secure-key" http://localhost:3000/api/collections
```

---

## 💻 Client Code Examples

### Python (Simple Script)
```python
import requests

API_URL = "http://localhost:3000/api"
API_KEY = "toondb-secure-key"
HEADERS = {"X-API-Key": API_KEY, "Content-Type": "text/plain"}

# 1. Save or Update Data
toon_data = """
name: Sara
role: Data Scientist
skills[2]: python,pytorch
"""
response = requests.post(f"{API_URL}/users/sara", data=toon_data, headers=HEADERS)
print("Save Status:", response.json())

# 2. Read Data
response = requests.get(f"{API_URL}/users/sara", headers=HEADERS)
print("\nData Received:\n", response.text)
```

### Python (Class Wrapper)
For a cleaner, object-oriented approach:
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

# Usage:
db = ToonDB()

# Save
db.save("myresume", "personal", "name: John\njob: Developer")

# Get
print(db.get("myresume", "personal"))

# Delete
# db.delete("myresume", "personal")
```

### Node.js
```javascript
const API_URL = 'http://localhost:3000/api';
const API_KEY = 'toondb-secure-key';

async function main() {
  const headers = { 'X-API-Key': API_KEY, 'Content-Type': 'text/plain' };

  // 1. Save or Update Data
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

  // 2. Read Data
  const readRes = await fetch(`${API_URL}/users/mike`, { headers });
  console.log('\nData Received:\n', await readRes.text());
}

main();
```

---

## 💾 Backup & Restore

### Backup
Download a JSON dump of the entire database:
```bash
curl -H "X-API-Key: toondb-secure-key" http://localhost:3000/api/backup > backup.json
```

### Restore
**Warning:** This will merge/overwrite existing data with the backup file.
```bash
curl -X POST http://localhost:3000/api/restore \
  -H "X-API-Key: toondb-secure-key" \
  -H "Content-Type: application/json" \
  -d @backup.json
```

---

## 📝 TOON Format Syntax
TOON is similar to YAML but simplified:

```text
# Simple Key-Value
title: Project Manager

# Simple Array
tags[3]: urgent,backend,api

# Nested Object
metadata:
  created: 2023-10-01
  author: admin

# Array of Objects (Table style)
users[2]{id,name}:
  1,Ali
  2,Sara
```
