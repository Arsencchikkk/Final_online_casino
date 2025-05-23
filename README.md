# 🎰 Online Casino

Online Casino is a fully functional web application where users can play Blackjack, manage their wallet balance, and enjoy a responsive and secure system. This project is built with Golang, gRPC, MongoDB, SMTP, NATS, and a lightweight vanilla JavaScript frontend.

## 🚀 Features

- 🃏 Play Blackjack (21)
- 👤 User registration and login with JWT authentication
- 💼 Wallet management (deposit and withdraw)
- 📧 Email verification via SMTP
- 💬 Event-driven communication with NATS
- 🛡 Secure endpoints with JWT and input validation
- 💻 Clean frontend using HTML/CSS/JS (no frameworks)

## 📦 Tech Stack

| Technology | Purpose                            |
|------------|------------------------------------|
| Go         | Backend and microservices logic    |
| gRPC       | Communication between services     |
| MongoDB    | Persistent storage (wallets, users)|
| SMTP       | Email confirmation system          |
| NATS       | Message-based communication        |
| HTML/CSS/JS| Frontend UI (no frameworks)        |
| Unit Test  | Unit testing (Mock and Integration)|


## 📂 Project Structure
casino/
│
├── game_service/         # gRPC service for Blackjack
├── wallet_service/       # gRPC service for wallet management (MongoDB)
├── user_service/         # Handles registration, login, SMTP, JWT
├── frontend/             # HTML, CSS, and JS files
│   ├── index.html
│   ├── game.html
│   ├── css/
│   └── js/
├── proto/                # .proto files for gRPC APIs
├── README.md
└── go.mod


## ⚙️ Getting Started

> ⚠ Make sure Go, MongoDB, and NATS are installed on your system.

1. Clone the repository:
   ```bash
   git clone https://github.com/your-username/online-casino.git
   cd online-casino

2. Install Go dependencies:
```bash
   go mod tidy

3. Run each service in its folder:
 • user_service
 • wallet_service
 • game_service

4. Open frontend/index.html in your browser.


🤝 Contributing

We welcome contributions! Feel free to open an issue or submit a pull request.
Project made by Arsen Bayakhmet, Abylaikhan Sekerbek, Danelya Maxutova.

📬 Contact

If you have any questions or suggestions:
 • Email: 231737@astanait.edu.kz
 • Telegram: @abylai_s7, @Lednik7lvl, @daniyuwwa


