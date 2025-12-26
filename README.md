# 🛡️ TEMPORARY ROOM  
## Secure & Concurrent Chat Service

**TEMPORARY ROOM** is a high-performance real-time chat application built with **Go**.  
It is designed to provide **secure, isolated, and ephemeral chat rooms**, where all client data is **automatically deleted after a predefined time**, ensuring privacy by design.

The system heavily relies on **advanced Go concurrency mechanisms** to achieve scalability, fault isolation, and strong security guarantees.

---

## 📱 Mobile Support & Compatibility

- **Android Support**  
  Fully supported on Android **phones and tablets** with high performance.

- **Platform Compatibility**  
  Supports **open platforms (Android)**.  
  ❌ iOS is not supported in the current version.

- **Responsive UI**  
  Optimized for touch-based devices with fast response time.

---

## 🏗️ Concurrency–Security Architecture

The core strength of this project lies in combining **Go Concurrency** with **Security principles**:

### 🔹 Isolation via Goroutines
- Each chat room runs in its own **Goroutine** (`room.run()`).
- Guarantees full **data isolation** between rooms.
- Prevents system-wide failure caused by a single room crash.

### 🔹 Safe Communication via Channels
- All communication between Goroutines is handled **exclusively using Go Channels**.
- Prevents **data races** without using complex locks.
- Ensures thread-safe message broadcasting and client registration.

### 🔹 DoS Protection (Slow Client Mitigation)
- Uses **non-blocking message delivery** (`select` with `default`).
- Automatically disconnects slow or malicious clients.
- Protects server resources from denial-of-service attacks.

### 🔹 Ephemeral Data Policy
- Chat rooms have a **limited lifetime** (e.g., 5 minutes).
- Once expired:
  - `CloseRoom` is triggered
  - Channels are closed
  - Room data is fully removed from memory
- Ensures **client data privacy** and secure cleanup.

---

## 📂 Project Structure

```plaintext
.
├── internal/
│   ├── chat/        # Room & Client logic (Goroutines / Channels)
│   ├── hub/         # Central room manager
│   └── websocket/   # WebSocket connection handlers
├── docs/            # Full documentation & team details (PDF)
├── main.go          # Application entry point
├── go.mod           # Dependency management
└── README.md        # Project documentation
