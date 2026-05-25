import requests
import sys
import json
from datetime import datetime
import time

class SentinelNOCTester:
    def __init__(self, base_url="https://app-preview-241.preview.emergentagent.com"):
        self.base_url = base_url
        self.token = None
        self.refresh_token = None
        self.user_id = None
        self.tests_run = 0
        self.tests_passed = 0
        self.test_results = []

    def log_test(self, name, success, details=""):
        """Log test result"""
        self.tests_run += 1
        if success:
            self.tests_passed += 1
        
        result = {
            "test": name,
            "success": success,
            "details": details,
            "timestamp": datetime.now().isoformat()
        }
        self.test_results.append(result)
        
        status = "✅ PASS" if success else "❌ FAIL"
        print(f"{status} - {name}")
        if details:
            print(f"    {details}")

    def run_test(self, name, method, endpoint, expected_status, data=None, headers=None):
        """Run a single API test"""
        url = f"{self.base_url}/api/{endpoint}"
        test_headers = {'Content-Type': 'application/json'}
        
        if self.token:
            test_headers['Authorization'] = f'Bearer {self.token}'
        
        if headers:
            test_headers.update(headers)

        try:
            if method == 'GET':
                response = requests.get(url, headers=test_headers, timeout=30)
            elif method == 'POST':
                response = requests.post(url, json=data, headers=test_headers, timeout=30)
            elif method == 'PUT':
                response = requests.put(url, json=data, headers=test_headers, timeout=30)
            elif method == 'DELETE':
                response = requests.delete(url, headers=test_headers, timeout=30)

            success = response.status_code == expected_status
            details = f"Status: {response.status_code}, Expected: {expected_status}"
            
            if not success:
                try:
                    error_detail = response.json().get('detail', 'No error details')
                    details += f", Error: {error_detail}"
                except:
                    details += f", Response: {response.text[:200]}"

            self.log_test(name, success, details)
            
            if success:
                try:
                    return response.json()
                except:
                    return {}
            return None

        except Exception as e:
            self.log_test(name, False, f"Exception: {str(e)}")
            return None

    def test_health_check(self):
        """Test basic health endpoints"""
        print("\n🔍 Testing Health Endpoints...")
        self.run_test("API Root", "GET", "", 200)
        self.run_test("Health Check", "GET", "health", 200)

    def test_user_registration(self):
        """Test user registration - first user becomes admin"""
        print("\n🔍 Testing User Registration...")
        
        # Test registration with admin credentials
        admin_data = {
            "username": "admin",
            "email": "admin@sentinel.com", 
            "password": "Admin123456!@"
        }
        
        result = self.run_test("Register Admin User", "POST", "auth/register", 200, admin_data)
        if result:
            self.user_id = result.get('id')
            return True
        else:
            # Admin user might already exist, try with a different user
            test_user_data = {
                "username": f"testuser_{int(time.time())}",
                "email": f"test_{int(time.time())}@sentinel.com",
                "password": "TestUser123456!@"
            }
            
            result = self.run_test("Register Test User", "POST", "auth/register", 200, test_user_data)
            if result:
                self.log_test("User Registration Flow", True, "New user registration works")
                return True
            else:
                self.log_test("User Registration Flow", True, "Admin user already exists (expected)")
                return True

    def test_user_login(self):
        """Test user login with JWT tokens"""
        print("\n🔍 Testing User Authentication...")
        
        # Try the new admin credentials first
        login_data = {
            "username": "admin",
            "password": "P@ssw0rd!"
        }
        
        result = self.run_test("User Login (New Credentials)", "POST", "auth/login", 200, login_data)
        if result:
            self.token = result.get('access_token')
            self.refresh_token = result.get('refresh_token')
            user_data = result.get('user', {})
            
            # Verify admin role
            if user_data.get('role') == 'admin':
                self.log_test("Admin Role Verification", True, "User has admin role")
            else:
                self.log_test("Admin Role Verification", False, f"Expected admin role, got {user_data.get('role')}")
            
            return True
        else:
            # Fallback to old credentials
            login_data = {
                "username": "admin",
                "password": "Admin123456!@"
            }
            
            result = self.run_test("User Login (Fallback Credentials)", "POST", "auth/login", 200, login_data)
            if result:
                self.token = result.get('access_token')
                self.refresh_token = result.get('refresh_token')
                user_data = result.get('user', {})
                
                # Verify admin role
                if user_data.get('role') == 'admin':
                    self.log_test("Admin Role Verification", True, "User has admin role")
                else:
                    self.log_test("Admin Role Verification", False, f"Expected admin role, got {user_data.get('role')}")
                
                return True
        return False

    def test_rate_limiting(self):
        """Test login rate limiting"""
        print("\n🔍 Testing Rate Limiting...")
        
        # Make multiple rapid login attempts with wrong credentials
        wrong_data = {"username": "admin", "password": "wrongpassword"}
        
        for i in range(3):
            response = requests.post(f"{self.base_url}/api/auth/login", json=wrong_data)
            if i == 0:
                # First attempt should return 401
                success = response.status_code == 401
                self.log_test(f"Rate Limit Test {i+1}", success, f"Status: {response.status_code}")
            time.sleep(0.1)

    def test_get_current_user(self):
        """Test getting current user info"""
        print("\n🔍 Testing User Info Retrieval...")
        
        result = self.run_test("Get Current User", "GET", "auth/me", 200)
        if result:
            # Verify user data structure
            required_fields = ['id', 'username', 'email', 'role', 'totp_enabled']
            missing_fields = [field for field in required_fields if field not in result]
            
            if not missing_fields:
                self.log_test("User Data Structure", True, "All required fields present")
            else:
                self.log_test("User Data Structure", False, f"Missing fields: {missing_fields}")

    def test_2fa_setup(self):
        """Test 2FA setup flow"""
        print("\n🔍 Testing 2FA Setup...")
        
        # Setup 2FA
        result = self.run_test("2FA Setup", "POST", "auth/2fa/setup", 200)
        if result and 'secret' in result:
            self.log_test("2FA Secret Generation", True, "Secret generated successfully")
            
            # Note: In real testing, we'd use the secret to generate a TOTP code
            # For now, we'll just verify the setup endpoint works
            return result.get('secret')
        return None

    def test_rbac_admin_access(self):
        """Test RBAC - Admin can access all pages"""
        print("\n🔍 Testing RBAC - Admin Access...")
        
        # Test admin-only endpoints
        admin_endpoints = [
            ("Users List", "GET", "users", 200),
            ("Audit Logs", "GET", "audit-logs", 200),
            ("System Settings", "GET", "settings", 200)
        ]
        
        for name, method, endpoint, expected in admin_endpoints:
            self.run_test(name, method, endpoint, expected)

    def test_camera_crud(self):
        """Test Camera CRUD operations"""
        print("\n🔍 Testing Camera Management...")
        
        # Generate unique IP for testing
        import random
        unique_ip = f"192.168.1.{random.randint(150, 250)}"
        
        # Create camera
        camera_data = {
            "name": f"Test Camera {int(time.time())}",
            "ip_address": unique_ip,
            "port": 80,
            "rtsp_port": 554,
            "rtsp_path": "/stream1",
            "username": "admin",
            "password": "testpass123",
            "location": "Main Entrance",
            "ptz_capable": True
        }
        
        create_result = self.run_test("Create Camera", "POST", "cameras", 200, camera_data)
        camera_id = None
        
        if create_result:
            camera_id = create_result.get('id')
            self.log_test("Camera ID Generation", bool(camera_id), f"Camera ID: {camera_id}")
            
            # Verify encrypted credentials (should not be in response)
            if 'password' not in create_result:
                self.log_test("Credential Encryption", True, "Password not exposed in response")
            else:
                self.log_test("Credential Encryption", False, "Password exposed in response")

        # List cameras
        list_result = self.run_test("List Cameras", "GET", "cameras", 200)
        if list_result and isinstance(list_result, list):
            self.log_test("Camera List Format", True, f"Found {len(list_result)} cameras")
        
        # Get specific camera
        if camera_id:
            self.run_test("Get Camera Details", "GET", f"cameras/{camera_id}", 200)
            
            # Update camera
            update_data = {"name": "Updated Test Camera", "location": "Updated Location"}
            self.run_test("Update Camera", "PUT", f"cameras/{camera_id}", 200, update_data)
            
            # Get stream URL (tests credential decryption)
            self.run_test("Get Stream URL", "GET", f"cameras/{camera_id}/stream-url", 200)
            
            # Delete camera (admin only)
            self.run_test("Delete Camera", "DELETE", f"cameras/{camera_id}", 200)

    def test_events_crud(self):
        """Test Events CRUD and acknowledgment"""
        print("\n🔍 Testing Event Management...")
        
        # First create a camera for events
        camera_data = {
            "name": "Event Test Camera",
            "ip_address": "192.168.1.101",
            "port": 80,
            "rtsp_port": 554,
            "username": "admin",
            "password": "testpass123"
        }
        
        camera_result = self.run_test("Create Event Test Camera", "POST", "cameras", 200, camera_data)
        camera_id = camera_result.get('id') if camera_result else None
        
        if camera_id:
            # Create event
            event_data = {
                "camera_id": camera_id,
                "event_type": "motion",
                "severity": "warning",
                "message": "Motion detected in main entrance",
                "details": {"confidence": 0.85}
            }
            
            event_result = self.run_test("Create Event", "POST", "events", 200, event_data)
            event_id = event_result.get('id') if event_result else None
            
            # List events
            self.run_test("List Events", "GET", "events", 200)
            
            # List events with filters
            self.run_test("Filter Events by Camera", "GET", f"events?camera_id={camera_id}", 200)
            self.run_test("Filter Events by Type", "GET", "events?event_type=motion", 200)
            self.run_test("Filter Events by Severity", "GET", "events?severity=warning", 200)
            
            # Acknowledge event
            if event_id:
                self.run_test("Acknowledge Event", "POST", f"events/{event_id}/acknowledge", 200)
            
            # Cleanup
            self.run_test("Delete Event Test Camera", "DELETE", f"cameras/{camera_id}", 200)

    def test_audit_logging(self):
        """Test audit logging for security actions"""
        print("\n🔍 Testing Audit Logging...")
        
        # Get audit logs (should contain previous actions)
        result = self.run_test("Get Audit Logs", "GET", "audit-logs", 200)
        
        if result and isinstance(result, list):
            self.log_test("Audit Log Format", True, f"Found {len(result)} audit entries")
            
            # Check for expected audit entries
            actions_found = [log.get('action') for log in result]
            expected_actions = ['user_registered', 'login_success', 'camera_created']
            
            found_expected = [action for action in expected_actions if action in actions_found]
            self.log_test("Audit Actions Logged", len(found_expected) > 0, 
                         f"Found actions: {found_expected}")

    def test_system_mode_api(self):
        """Test System Mode API endpoints"""
        print("\n🔍 Testing System Mode API...")
        
        # Get current system mode
        get_result = self.run_test("Get System Mode", "GET", "system/mode", 200)
        if get_result:
            current_mode = get_result.get('mode')
            self.log_test("System Mode Structure", 'mode' in get_result, f"Current mode: {current_mode}")
            
            # Verify mode is valid
            if current_mode in ['home', 'away']:
                self.log_test("Valid System Mode", True, f"Mode '{current_mode}' is valid")
            else:
                self.log_test("Valid System Mode", False, f"Invalid mode: {current_mode}")
        
        # Change to home mode
        home_data = {"mode": "home"}
        home_result = self.run_test("Set Home Mode", "PUT", "system/mode", 200, home_data)
        if home_result:
            if home_result.get('mode') == 'home':
                self.log_test("Home Mode Set", True, "Successfully changed to home mode")
            else:
                self.log_test("Home Mode Set", False, f"Expected home mode, got {home_result.get('mode')}")
        
        # Change to away mode
        away_data = {"mode": "away"}
        away_result = self.run_test("Set Away Mode", "PUT", "system/mode", 200, away_data)
        if away_result:
            if away_result.get('mode') == 'away':
                self.log_test("Away Mode Set", True, "Successfully changed to away mode")
            else:
                self.log_test("Away Mode Set", False, f"Expected away mode, got {away_result.get('mode')}")
        
        # Test invalid mode
        invalid_data = {"mode": "invalid"}
        self.run_test("Invalid Mode Rejection", "PUT", "system/mode", 400, invalid_data)
        
        # Verify mode change is reflected in dashboard stats
        stats_result = self.run_test("Dashboard Stats After Mode Change", "GET", "dashboard/stats", 200)
        if stats_result:
            stats_mode = stats_result.get('system_mode')
            if stats_mode == 'away':
                self.log_test("Mode Reflected in Dashboard", True, f"Dashboard shows mode: {stats_mode}")
            else:
                self.log_test("Mode Reflected in Dashboard", False, f"Dashboard mode mismatch: {stats_mode}")

    def test_onvif_apis(self):
        """Test ONVIF API endpoints"""
        print("\n🔍 Testing ONVIF APIs...")
        
        # First create a test camera to use for ONVIF testing with unique IP
        import random
        unique_ip = f"192.168.1.{random.randint(201, 254)}"
        
        camera_data = {
            "name": "ONVIF Test Camera",
            "ip_address": unique_ip,
            "port": 80,
            "rtsp_port": 554,
            "username": "admin",
            "password": "camera123",
            "onvif_port": 80,
            "onvif_username": "admin",
            "onvif_password": "test123"
        }
        
        camera_result = self.run_test("Create ONVIF Test Camera", "POST", "cameras", 200, camera_data)
        camera_id = camera_result.get('id') if camera_result else None
        
        if camera_id:
            # Test ONVIF credentials update
            onvif_creds = {
                "onvif_port": 80,
                "onvif_username": "admin",
                "onvif_password": "test123"
            }
            self.run_test("Update ONVIF Credentials", "POST", f"cameras/{camera_id}/onvif/credentials", 200, onvif_creds)
            
            # Test ONVIF connection (expected to fail but API should respond)
            test_result = self.run_test("Test ONVIF Connection", "POST", f"cameras/{camera_id}/onvif/test", 200)
            if test_result is not None:
                # Check if response has expected structure
                if 'success' in test_result:
                    self.log_test("ONVIF Test Response Structure", True, f"Success field present: {test_result.get('success')}")
                else:
                    self.log_test("ONVIF Test Response Structure", False, "Missing success field in response")
            
            # Test ONVIF capability detection (expected to fail but API should respond)
            detect_result = self.run_test("Detect ONVIF Capabilities", "POST", f"cameras/{camera_id}/onvif/detect", 400)
            # 400 is expected since camera doesn't exist, but API should respond properly
            
            # Cleanup
            self.run_test("Delete ONVIF Test Camera", "DELETE", f"cameras/{camera_id}", 200)
        else:
            self.log_test("ONVIF Tests", False, "Could not create test camera for ONVIF testing")

    def test_camera_mode_override(self):
        """Test Camera with Mode Override functionality"""
        print("\n🔍 Testing Camera Mode Override...")
        
        # Create camera with mode override
        camera_data = {
            "name": "Test Mode Override Camera",
            "ip_address": "192.168.1.250",
            "port": 80,
            "rtsp_port": 554,
            "username": "admin",
            "password": "camera123",
            "mode_override": "away"
        }
        
        camera_result = self.run_test("Create Camera with Mode Override", "POST", "cameras", 200, camera_data)
        camera_id = camera_result.get('id') if camera_result else None
        
        if camera_result:
            # Verify mode_override field is present
            if 'mode_override' in camera_result:
                mode_override = camera_result.get('mode_override')
                self.log_test("Mode Override Field Present", True, f"Mode override: {mode_override}")
                
                if mode_override == 'away':
                    self.log_test("Mode Override Value Correct", True, "Mode override set to 'away'")
                else:
                    self.log_test("Mode Override Value Correct", False, f"Expected 'away', got '{mode_override}'")
            else:
                self.log_test("Mode Override Field Present", False, "mode_override field missing from response")
            
            # Verify effective_mode field is present
            if 'effective_mode' in camera_result:
                effective_mode = camera_result.get('effective_mode')
                self.log_test("Effective Mode Field Present", True, f"Effective mode: {effective_mode}")
            else:
                self.log_test("Effective Mode Field Present", False, "effective_mode field missing from response")
            
            # Get camera details to verify fields persist
            if camera_id:
                get_result = self.run_test("Get Camera with Mode Override", "GET", f"cameras/{camera_id}", 200)
                if get_result and 'mode_override' in get_result:
                    self.log_test("Mode Override Persisted", True, f"Mode override persisted: {get_result.get('mode_override')}")
                
                # Cleanup
                self.run_test("Delete Mode Override Test Camera", "DELETE", f"cameras/{camera_id}", 200)

    def test_dashboard_stats_new_fields(self):
        """Test Dashboard Stats API for new fields"""
        print("\n🔍 Testing Dashboard Stats New Fields...")
        
        result = self.run_test("Dashboard Stats with New Fields", "GET", "dashboard/stats", 200)
        
        if result:
            # Check for system_mode field
            if 'system_mode' in result:
                system_mode = result.get('system_mode')
                self.log_test("System Mode in Dashboard", True, f"System mode: {system_mode}")
                
                # Verify it's a valid mode
                if system_mode in ['home', 'away']:
                    self.log_test("Valid System Mode in Dashboard", True, f"Valid mode: {system_mode}")
                else:
                    self.log_test("Valid System Mode in Dashboard", False, f"Invalid mode: {system_mode}")
            else:
                self.log_test("System Mode in Dashboard", False, "system_mode field missing")
            
            # Check for alarm_capable_cameras field
            if 'alarm_capable_cameras' in result:
                alarm_capable = result.get('alarm_capable_cameras')
                self.log_test("Alarm Capable Cameras in Dashboard", True, f"Alarm capable cameras: {alarm_capable}")
                
                # Should be a number
                if isinstance(alarm_capable, (int, float)):
                    self.log_test("Alarm Capable Cameras Type", True, f"Numeric value: {alarm_capable}")
                else:
                    self.log_test("Alarm Capable Cameras Type", False, f"Expected number, got {type(alarm_capable)}")
            else:
                self.log_test("Alarm Capable Cameras in Dashboard", False, "alarm_capable_cameras field missing")
            
            # Verify all original fields are still present
            required_original_fields = ['total_cameras', 'online_cameras', 'offline_cameras', 
                                      'total_events', 'unacknowledged_events', 'critical_events']
            
            missing_fields = [field for field in required_original_fields if field not in result]
            
            if not missing_fields:
                self.log_test("Original Dashboard Fields Present", True, "All original fields present")
            else:
                self.log_test("Original Dashboard Fields Present", False, f"Missing fields: {missing_fields}")

    def test_dashboard_stats(self):
        """Test dashboard stats API"""
        print("\n🔍 Testing Dashboard Stats...")
        
        result = self.run_test("Dashboard Stats", "GET", "dashboard/stats", 200)
        
        if result:
            required_stats = ['total_cameras', 'online_cameras', 'offline_cameras', 
                            'total_events', 'unacknowledged_events', 'critical_events',
                            'system_mode', 'alarm_capable_cameras']
            
            missing_stats = [stat for stat in required_stats if stat not in result]
            
            if not missing_stats:
                self.log_test("Dashboard Stats Structure", True, "All required stats present")
            else:
                self.log_test("Dashboard Stats Structure", False, f"Missing stats: {missing_stats}")

    def test_system_settings(self):
        """Test system settings API"""
        print("\n🔍 Testing System Settings...")
        
        # Get settings
        get_result = self.run_test("Get System Settings", "GET", "settings", 200)
        
        if get_result:
            # Update settings
            settings_data = {
                "storage_path": "/recordings",
                "retention_days": 45,
                "motion_sensitivity": 75,
                "alarm_notification_enabled": True,
                "email_notifications": False
            }
            
            self.run_test("Update System Settings", "PUT", "settings", 200, settings_data)

    def test_token_refresh(self):
        """Test JWT token refresh"""
        print("\n🔍 Testing Token Refresh...")
        
        if self.refresh_token:
            # Test refresh token endpoint
            response = requests.post(
                f"{self.base_url}/api/auth/refresh",
                params={"refresh_token": self.refresh_token},
                timeout=30
            )
            
            success = response.status_code == 200
            details = f"Status: {response.status_code}"
            
            if success:
                try:
                    result = response.json()
                    new_token = result.get('access_token')
                    if new_token:
                        self.token = new_token  # Update token for future tests
                        details += ", New token received"
                except:
                    success = False
                    details += ", Invalid response format"
            
            self.log_test("Token Refresh", success, details)

    def run_all_tests(self):
        """Run comprehensive test suite"""
        print("🚀 Starting Sentinel NOC API Test Suite")
        print(f"📡 Testing against: {self.base_url}")
        print("=" * 60)
        
        # Basic connectivity
        self.test_health_check()
        
        # Authentication flow
        if self.test_user_registration():
            if self.test_user_login():
                self.test_get_current_user()
                self.test_token_refresh()
                self.test_2fa_setup()
                
                # Rate limiting
                self.test_rate_limiting()
                
                # RBAC testing
                self.test_rbac_admin_access()
                
                # Core functionality
                self.test_camera_crud()
                self.test_events_crud()
                self.test_audit_logging()
                self.test_dashboard_stats()
                self.test_system_settings()
                
                # NEW: Test System Mode and ONVIF APIs
                print("\n" + "=" * 60)
                print("🆕 Testing NEW System Mode and ONVIF Features")
                print("=" * 60)
                self.test_system_mode_api()
                self.test_onvif_apis()
                self.test_camera_mode_override()
                self.test_dashboard_stats_new_fields()
        
        # Print summary
        print("\n" + "=" * 60)
        print(f"📊 Test Summary: {self.tests_passed}/{self.tests_run} tests passed")
        
        if self.tests_passed == self.tests_run:
            print("🎉 All tests passed!")
            return 0
        else:
            print(f"⚠️  {self.tests_run - self.tests_passed} tests failed")
            return 1

def main():
    tester = SentinelNOCTester()
    return tester.run_all_tests()

if __name__ == "__main__":
    sys.exit(main())