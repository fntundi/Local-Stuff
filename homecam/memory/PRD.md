# Sentinel NOC - Security Camera Network Operations Center

## Project Overview
A comprehensive Security Camera Network Operations Center (NOC) / Video Management System built with React, FastAPI, and MongoDB. The system implements a zero-trust security model with encrypted credentials, JWT authentication, TOTP 2FA, and role-based access control.

## Original Problem Statement
Build a secure camera management system with:
- ONVIF camera discovery and RTSP streaming
- Zero-trust security model treating all devices as untrusted
- Full RBAC (Admin, Security Operator, Viewer)
- JWT authentication with TOTP 2FA
- Encrypted camera credentials
- Audit logging
- Motion/alarm event system

## User Choices
- Real ONVIF/RTSP camera integration
- JWT-based custom auth with TOTP 2FA (pyotp)
- MongoDB database
- Dark theme ("Sentinel OS" design)

## User Personas

### Admin
- Full system access
- User management, RBAC configuration
- System settings (storage, retention, LDAP)
- Audit log access
- Camera deletion

### Security Operator
- Live viewing, PTZ control
- Alarm acknowledgment
- Recording playback
- Camera configuration (add/edit)
- Event monitoring

### Viewer
- Live viewing only
- No configuration changes
- No recording deletion
- No user management

## Core Requirements (Static)

### Authentication & Security
- [x] JWT access/refresh tokens
- [x] TOTP 2FA with pyotp
- [x] Password hashing (bcrypt)
- [x] Rate limiting per IP
- [x] Account lockout after 5 failed attempts
- [x] Encrypted camera credentials (Fernet)
- [x] Session management

### RBAC
- [x] Admin role (full access)
- [x] Security Operator role (operations)
- [x] Viewer role (read-only)
- [x] Server-side authorization enforcement

### Camera Management
- [x] Camera CRUD operations
- [x] IP/port/RTSP configuration
- [x] Credential encryption at rest
- [x] PTZ capability flag
- [x] Status tracking (online/offline)

### Event System
- [x] Event creation (motion, alarm, connection)
- [x] Severity levels (info, warning, critical)
- [x] Event filtering
- [x] Event acknowledgment

### Audit Logging
- [x] All security actions logged
- [x] User actions tracked
- [x] IP address logging
- [x] Filterable audit log viewer

## What's Been Implemented (MVP)

### Date: 2026-02-24

#### Backend (FastAPI)
- Complete authentication system with JWT + 2FA
- RBAC middleware with role checking
- Camera CRUD with encrypted credentials
- Events/Alarms system
- Audit logging
- Dashboard stats API
- System settings API
- Rate limiting and account lockout

#### Frontend (React)
- Login page with 2FA support
- Dashboard with camera grid (1x1, 2x2, 3x3, 4x4)
- Camera management page
- Camera detail page with PTZ controls
- Events page with filtering
- User management (Admin)
- Settings page (Admin)
- Audit logs page (Admin)
- Responsive sidebar navigation

#### Design System ("Sentinel OS")
- Dark theme with cyan accent
- JetBrains Mono for data/mono text
- Chivo for headings
- Scanline effect on video feeds
- Glass morphism cards
- Status badges (live, online, offline)

## Prioritized Backlog

### P0 (Critical)
- [x] Authentication system
- [x] Camera management
- [x] RBAC enforcement
- [x] Audit logging

### P1 (High)
- [ ] Real RTSP stream playback (HLS transcoding)
- [ ] ONVIF camera discovery
- [ ] Real motion detection
- [ ] Recording system with NAS storage

### P2 (Medium)
- [ ] LDAP integration
- [ ] Email notifications (SMTP)
- [ ] Webhook integration
- [ ] Multi-site support

### P3 (Low)
- [ ] Hardware acceleration detection
- [ ] Export recordings
- [ ] Mobile app support
- [ ] Advanced analytics

## Next Tasks

1. **RTSP Stream Integration**
   - Implement HLS transcoding with FFmpeg
   - WebRTC for low-latency viewing
   - Real camera stream playback in browser

2. **ONVIF Discovery**
   - WS-Discovery implementation
   - Auto-populate camera details
   - PTZ control via ONVIF

3. **Recording System**
   - Continuous recording
   - Motion-triggered recording
   - NAS storage integration
   - Recording playback timeline

4. **Motion Detection**
   - Server-side motion analysis
   - Configurable sensitivity
   - Alert generation

## Technical Stack

### Backend
- FastAPI 0.110.1
- MongoDB (Motor async driver)
- PyJWT for authentication
- PyOTP for 2FA
- Cryptography (Fernet) for encryption
- Bcrypt for password hashing

### Frontend
- React 19
- React Router 7
- Tailwind CSS
- shadcn/ui components
- Axios for API calls
- Sonner for toasts

### Security Features
- JWT with 30-min access tokens
- 7-day refresh tokens
- TOTP 2FA (required for admin)
- Encrypted credentials at rest
- Rate limiting (100/min per IP)
- Account lockout (15 min after 5 failures)
- Audit logging all actions

## API Endpoints

### Auth
- POST /api/auth/register
- POST /api/auth/login
- POST /api/auth/refresh
- GET /api/auth/me
- POST /api/auth/2fa/setup
- POST /api/auth/2fa/verify

### Cameras
- GET/POST /api/cameras
- GET/PUT/DELETE /api/cameras/{id}
- GET /api/cameras/{id}/stream-url

### Events
- GET/POST /api/events
- POST /api/events/{id}/acknowledge

### Admin
- GET /api/users
- PUT /api/users/{id}/role
- DELETE /api/users/{id}
- GET/PUT /api/settings
- GET /api/audit-logs
- GET /api/dashboard/stats
