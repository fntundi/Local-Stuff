#!/usr/bin/env python3
"""
Sentinel NOC Backend API Test Suite
Tests the specific endpoints requested for the Go backend
"""
import requests
import json
import sys
from datetime import datetime

class SentinelNOCAPITester:
    def __init__(self):
        # Use the production URL from frontend/.env
        self.base_url = "https://app-preview-241.preview.emergentagent.com"
        self.access_token = None
        self.refresh_token = None
        self.camera_id = None
        self.tests_passed = 0
        self.tests_failed = 0
        self.test_results = []

    def log_result(self, test_name, success, details="", response_data=None):
        """Log test result with details"""
        if success:
            self.tests_passed += 1
            print(f"✅ {test_name}")
        else:
            self.tests_failed += 1
            print(f"❌ {test_name}")
        
        if details:
            print(f"   {details}")
        
        if response_data and isinstance(response_data, dict):
            print(f"   Response keys: {list(response_data.keys())}")
        
        self.test_results.append({
            "test": test_name,
            "success": success,
            "details": details,
            "timestamp": datetime.now().isoformat()
        })

    def make_request(self, method, endpoint, data=None, use_auth=False):
        """Make HTTP request to API"""
        url = f"{self.base_url}/api/{endpoint}"
        headers = {"Content-Type": "application/json"}
        
        if use_auth and self.access_token:
            headers["Authorization"] = f"Bearer {self.access_token}"
        
        try:
            if method == "GET":
                response = requests.get(url, headers=headers, timeout=30)
            elif method == "POST":
                response = requests.post(url, json=data, headers=headers, timeout=30)
            elif method == "PUT":
                response = requests.put(url, json=data, headers=headers, timeout=30)
            elif method == "DELETE":
                response = requests.delete(url, headers=headers, timeout=30)
            
            return response
        except Exception as e:
            print(f"Request failed: {e}")
            return None

    def test_health_check(self):
        """Test 1: Health Check Endpoint"""
        print("\n🔍 Testing Health Check...")
        
        response = self.make_request("GET", "health")
        if response and response.status_code == 200:
            try:
                data = response.json()
                self.log_result("GET /api/health", True, f"Status: {response.status_code}", data)
                return True
            except:
                self.log_result("GET /api/health", False, "Invalid JSON response")
        else:
            status = response.status_code if response else "No response"
            self.log_result("GET /api/health", False, f"Status: {status}")
        return False

    def test_authentication(self):
        """Test 2: Authentication Flow"""
        print("\n🔍 Testing Authentication...")
        
        # Test login with specified credentials
        login_data = {
            "username": "admin",
            "password": "P@ssw0rd!"
        }
        
        response = self.make_request("POST", "auth/login", login_data)
        if response and response.status_code == 200:
            try:
                data = response.json()
                self.access_token = data.get("access_token")
                self.refresh_token = data.get("refresh_token")
                user_data = data.get("user", {})
                
                # Verify response structure
                required_fields = ["access_token", "refresh_token", "user"]
                missing_fields = [field for field in required_fields if field not in data]
                
                if not missing_fields and self.access_token:
                    self.log_result("POST /api/auth/login", True, 
                                  f"Login successful, user role: {user_data.get('role')}", data)
                    return True
                else:
                    self.log_result("POST /api/auth/login", False, 
                                  f"Missing fields: {missing_fields}")
            except Exception as e:
                self.log_result("POST /api/auth/login", False, f"JSON parse error: {e}")
        else:
            status = response.status_code if response else "No response"
            error_msg = ""
            if response:
                try:
                    error_data = response.json()
                    error_msg = error_data.get("detail", "")
                except:
                    error_msg = response.text[:100]
            self.log_result("POST /api/auth/login", False, f"Status: {status}, Error: {error_msg}")
        return False

    def test_token_refresh(self):
        """Test 3: Token Refresh"""
        print("\n🔍 Testing Token Refresh...")
        
        if not self.refresh_token:
            self.log_result("POST /api/auth/refresh", False, "No refresh token available")
            return False
        
        # Test refresh token endpoint
        response = self.make_request("POST", f"auth/refresh?refresh_token={self.refresh_token}")
        if response and response.status_code == 200:
            try:
                data = response.json()
                new_access_token = data.get("access_token")
                if new_access_token:
                    old_token = self.access_token[:20] + "..." if self.access_token else "None"
                    new_token = new_access_token[:20] + "..."
                    self.access_token = new_access_token  # Update for future tests
                    self.log_result("POST /api/auth/refresh", True, 
                                  f"Token refreshed successfully", data)
                    return True
                else:
                    self.log_result("POST /api/auth/refresh", False, "No access_token in response")
            except Exception as e:
                self.log_result("POST /api/auth/refresh", False, f"JSON parse error: {e}")
        else:
            status = response.status_code if response else "No response"
            self.log_result("POST /api/auth/refresh", False, f"Status: {status}")
        return False

    def test_get_current_user(self):
        """Test 4: Get Current User Info"""
        print("\n🔍 Testing Get Current User...")
        
        if not self.access_token:
            self.log_result("GET /api/auth/me", False, "No access token available")
            return False
        
        response = self.make_request("GET", "auth/me", use_auth=True)
        if response and response.status_code == 200:
            try:
                data = response.json()
                required_fields = ["id", "username", "email", "role"]
                missing_fields = [field for field in required_fields if field not in data]
                
                if not missing_fields:
                    self.log_result("GET /api/auth/me", True, 
                                  f"User info retrieved, role: {data.get('role')}", data)
                    return True
                else:
                    self.log_result("GET /api/auth/me", False, f"Missing fields: {missing_fields}")
            except Exception as e:
                self.log_result("GET /api/auth/me", False, f"JSON parse error: {e}")
        else:
            status = response.status_code if response else "No response"
            self.log_result("GET /api/auth/me", False, f"Status: {status}")
        return False

    def test_cameras_api(self):
        """Test 5: Cameras API"""
        print("\n🔍 Testing Cameras API...")
        
        if not self.access_token:
            self.log_result("Cameras API", False, "No access token available")
            return False
        
        # Test GET /api/cameras - list all cameras
        response = self.make_request("GET", "cameras", use_auth=True)
        if response and response.status_code == 200:
            try:
                cameras = response.json()
                self.log_result("GET /api/cameras", True, 
                              f"Retrieved {len(cameras)} cameras", {"count": len(cameras)})
            except Exception as e:
                self.log_result("GET /api/cameras", False, f"JSON parse error: {e}")
        else:
            status = response.status_code if response else "No response"
            self.log_result("GET /api/cameras", False, f"Status: {status}")
        
        # Test POST /api/cameras - create a camera
        camera_data = {
            "name": "Front Entrance",
            "ip_address": "192.168.1.100",
            "port": 80,
            "rtsp_port": 554,
            "username": "admin",
            "password": "camera123"
        }
        
        response = self.make_request("POST", "cameras", camera_data, use_auth=True)
        if response and response.status_code == 200:
            try:
                data = response.json()
                self.camera_id = data.get("id")
                self.log_result("POST /api/cameras", True, 
                              f"Camera created with ID: {self.camera_id}", data)
                
                # Test GET /api/cameras/{id} - get specific camera
                if self.camera_id:
                    response = self.make_request("GET", f"cameras/{self.camera_id}", use_auth=True)
                    if response and response.status_code == 200:
                        try:
                            camera_data = response.json()
                            self.log_result("GET /api/cameras/{id}", True, 
                                          f"Retrieved camera: {camera_data.get('name')}", camera_data)
                        except Exception as e:
                            self.log_result("GET /api/cameras/{id}", False, f"JSON parse error: {e}")
                    else:
                        status = response.status_code if response else "No response"
                        self.log_result("GET /api/cameras/{id}", False, f"Status: {status}")
                
            except Exception as e:
                self.log_result("POST /api/cameras", False, f"JSON parse error: {e}")
        else:
            status = response.status_code if response else "No response"
            error_msg = ""
            if response:
                try:
                    error_data = response.json()
                    error_msg = error_data.get("detail", "")
                except:
                    error_msg = response.text[:100]
            self.log_result("POST /api/cameras", False, f"Status: {status}, Error: {error_msg}")

    def test_events_api(self):
        """Test 6: Events API"""
        print("\n🔍 Testing Events API...")
        
        if not self.access_token:
            self.log_result("Events API", False, "No access token available")
            return False
        
        # Test GET /api/events - list events
        response = self.make_request("GET", "events", use_auth=True)
        if response and response.status_code == 200:
            try:
                events = response.json()
                self.log_result("GET /api/events", True, 
                              f"Retrieved {len(events)} events", {"count": len(events)})
            except Exception as e:
                self.log_result("GET /api/events", False, f"JSON parse error: {e}")
        else:
            status = response.status_code if response else "No response"
            self.log_result("GET /api/events", False, f"Status: {status}")
        
        # Test POST /api/events - create event (if we have a camera)
        if self.camera_id:
            event_data = {
                "camera_id": self.camera_id,
                "event_type": "motion",
                "severity": "warning",
                "message": "Motion detected at front entrance"
            }
            
            response = self.make_request("POST", "events", event_data, use_auth=True)
            if response and response.status_code == 200:
                try:
                    data = response.json()
                    event_id = data.get("id")
                    self.log_result("POST /api/events", True, 
                                  f"Event created with ID: {event_id}", data)
                except Exception as e:
                    self.log_result("POST /api/events", False, f"JSON parse error: {e}")
            else:
                status = response.status_code if response else "No response"
                error_msg = ""
                if response:
                    try:
                        error_data = response.json()
                        error_msg = error_data.get("detail", "")
                    except:
                        error_msg = response.text[:100]
                self.log_result("POST /api/events", False, f"Status: {status}, Error: {error_msg}")
        else:
            self.log_result("POST /api/events", False, "No camera ID available for event creation")

    def test_dashboard_stats(self):
        """Test 7: Dashboard Stats"""
        print("\n🔍 Testing Dashboard Stats...")
        
        if not self.access_token:
            self.log_result("Dashboard Stats", False, "No access token available")
            return False
        
        response = self.make_request("GET", "dashboard/stats", use_auth=True)
        if response and response.status_code == 200:
            try:
                data = response.json()
                expected_fields = ["total_cameras", "total_events"]
                present_fields = [field for field in expected_fields if field in data]
                
                self.log_result("GET /api/dashboard/stats", True, 
                              f"Stats retrieved with fields: {list(data.keys())}", data)
            except Exception as e:
                self.log_result("GET /api/dashboard/stats", False, f"JSON parse error: {e}")
        else:
            status = response.status_code if response else "No response"
            self.log_result("GET /api/dashboard/stats", False, f"Status: {status}")

    def run_all_tests(self):
        """Run all API tests in sequence"""
        print("🚀 Starting Sentinel NOC API Test Suite")
        print(f"📡 Testing against: {self.base_url}")
        print("=" * 60)
        
        # Run tests in order
        self.test_health_check()
        
        if self.test_authentication():
            self.test_token_refresh()
            self.test_get_current_user()
            self.test_cameras_api()
            self.test_events_api()
            self.test_dashboard_stats()
        else:
            print("❌ Authentication failed - skipping protected endpoint tests")
        
        # Print summary
        print("\n" + "=" * 60)
        print(f"📊 Test Results: {self.tests_passed} passed, {self.tests_failed} failed")
        
        if self.tests_failed == 0:
            print("🎉 All tests passed!")
            return True
        else:
            print(f"⚠️  {self.tests_failed} tests failed")
            return False

def main():
    """Main test runner"""
    tester = SentinelNOCAPITester()
    success = tester.run_all_tests()
    return 0 if success else 1

if __name__ == "__main__":
    sys.exit(main())