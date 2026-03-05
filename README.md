# 🏦 Nexora Bank | AI-Powered Core Banking with TigerBeetle

Nexora is a modern banking platform engineered for high availability and financial consistency. It leverages **TigerBeetle** as its distributed ledger technology (DLT) and **AI (OpenAI/Gemini)** for asset management through natural language processing.

---

## 1. Technical Architecture
The platform is built on a microservices architecture orchestrated by **Docker**:

* **Frontend**: Svelte + Tailwind CSS (Reactive SPA interface).
* **Backend**: Go (Gin Gonic) - Business logic engine and REST API.
* **Core Ledger**: [TigerBeetle](https://tigerbeetle.com/) - Specialized financial database (capable of up to 1M transactions per second).
* **Database**: PostgreSQL - User metadata and profile persistence.
* **AI Engine**: LLM integration via *Function Calling* for transaction execution and natural language queries.

---

## 2. UI/UX Requirements Implemented
* ✅ **Responsive Design**: Interface optimized for mobile and desktop devices.
* ✅ **Full Reactivity**: Balances update in real-time via *Svelte Stores* without page reloads.
* ✅ **User Feedback**: Elegant *Toast* notification system for success confirmation and error reporting.
* ✅ **Loading States**: Integrated loading spinners in buttons and chat to enhance perceived speed.
* ✅ **Form Validation**: Strict control over minimum amounts and data types on both client and server sides.

---

## 3. Key Features

### 🧠 AI Financial Assistant
Execute complex operations directly through the chat using natural language:
* *"Deposit 100 dollars into my savings account"*
* *"What is the transaction history of my checking account?"*
* *"Transfer 50 USD to account 4001-0002-1000"*

### ⚡ High-Performance Ledger
Nexora separates concerns: **Metadata** resides in Postgres, while the **Real-Time Balance** is managed by TigerBeetle, ensuring zero race conditions during high-concurrency transactions.

---

## 4. Installation & Deployment (Docker)
Follow these steps to spin up the entire environment in under 5 minutes:

### Step 1: Clone the Repository
```bash
git clone [https://github.com/your-user/nexora-bank.git](https://github.com/your-user/nexora-bank.git)
cd nexora-bank

DB_HOST=postgres
DB_USER=nexora_user
DB_PASSWORD=nexora_pass
DB_NAME=nexora_db
OPENAI_API_KEY=your_api_key_here
TIGERBEETLE_REPLICA_ADDRESSES=3000

docker-compose up --build

Method,Route,Description
POST,/api/accounts/create,Opens a new TigerBeetle account and links it to the user.
GET,/api/balance,Fetches consolidated real-time balances.
POST,/api/transfer,Executes transfers between internal TigerBeetle IDs.
POST,/api/chat,AI Assistant endpoint with Function Calling capabilities.
GET,/api/history,Retrieves detailed transfer history from the Ledger.

7. Repository
Official Link: https://github.com/your-user/nexora-bank