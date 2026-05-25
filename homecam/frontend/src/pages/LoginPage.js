import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "@/contexts/AuthContext";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { InputOTP, InputOTPGroup, InputOTPSlot } from "@/components/ui/input-otp";
import { toast } from "sonner";
import { Shield, Eye, EyeOff, AlertTriangle } from "lucide-react";

export default function LoginPage() {
  const navigate = useNavigate();
  const { login, register } = useAuth();
  
  // Login state
  const [loginUsername, setLoginUsername] = useState("");
  const [loginPassword, setLoginPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [requires2FA, setRequires2FA] = useState(false);
  const [totpCode, setTotpCode] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  
  // Register state
  const [regUsername, setRegUsername] = useState("");
  const [regEmail, setRegEmail] = useState("");
  const [regPassword, setRegPassword] = useState("");
  const [regConfirmPassword, setRegConfirmPassword] = useState("");

  const handleLogin = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    
    try {
      const result = await login(loginUsername, loginPassword, requires2FA ? totpCode : null);
      
      if (result.requires2FA) {
        setRequires2FA(true);
        toast.info("2FA Required", { description: "Enter your authentication code" });
      } else if (result.success) {
        toast.success("Login Successful", { description: "Welcome to Sentinel NOC" });
        navigate("/dashboard");
      }
    } catch (error) {
      const message = error.response?.data?.detail || "Login failed";
      toast.error("Authentication Failed", { description: message });
      if (requires2FA) {
        setTotpCode("");
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleRegister = async (e) => {
    e.preventDefault();
    
    if (regPassword !== regConfirmPassword) {
      toast.error("Passwords don't match");
      return;
    }
    
    if (regPassword.length < 12) {
      toast.error("Password must be at least 12 characters");
      return;
    }
    
    setIsLoading(true);
    
    try {
      await register(regUsername, regEmail, regPassword);
      toast.success("Registration Successful", { description: "You can now log in" });
      setRegUsername("");
      setRegEmail("");
      setRegPassword("");
      setRegConfirmPassword("");
    } catch (error) {
      const message = error.response?.data?.detail || "Registration failed";
      toast.error("Registration Failed", { description: message });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-[#09090b] flex items-center justify-center p-4">
      {/* Background Grid */}
      <div className="absolute inset-0 bg-[linear-gradient(rgba(39,39,42,0.3)_1px,transparent_1px),linear-gradient(90deg,rgba(39,39,42,0.3)_1px,transparent_1px)] bg-[size:50px_50px]" />
      
      <div className="relative z-10 w-full max-w-md">
        {/* Logo */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-cyan-500/10 rounded-sm border border-cyan-500/30 mb-4">
            <Shield className="w-10 h-10 text-cyan-500" />
          </div>
          <h1 className="text-3xl font-bold text-white tracking-tight uppercase font-['Chivo']">
            SENTINEL
          </h1>
          <p className="text-sm text-zinc-500 font-mono uppercase tracking-wider mt-1">
            Network Operations Center
          </p>
        </div>

        <Card className="bg-zinc-900/80 border-zinc-800 backdrop-blur-sm" data-testid="login-card">
          <Tabs defaultValue="login" className="w-full">
            <TabsList className="grid w-full grid-cols-2 bg-zinc-950 p-1 rounded-sm">
              <TabsTrigger 
                value="login" 
                className="data-[state=active]:bg-cyan-500 data-[state=active]:text-black font-medium uppercase tracking-wide text-sm"
                data-testid="login-tab"
              >
                Login
              </TabsTrigger>
              <TabsTrigger 
                value="register" 
                className="data-[state=active]:bg-cyan-500 data-[state=active]:text-black font-medium uppercase tracking-wide text-sm"
                data-testid="register-tab"
              >
                Register
              </TabsTrigger>
            </TabsList>

            <TabsContent value="login">
              <CardHeader className="pb-4">
                <CardTitle className="text-xl font-bold text-white uppercase tracking-tight">
                  {requires2FA ? "2FA Verification" : "Access Terminal"}
                </CardTitle>
                <CardDescription className="text-zinc-500 font-mono text-xs">
                  {requires2FA 
                    ? "Enter the 6-digit code from your authenticator app" 
                    : "Enter your credentials to access the system"}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <form onSubmit={handleLogin} className="space-y-4">
                  {!requires2FA ? (
                    <>
                      <div className="space-y-2">
                        <Label htmlFor="username" className="text-zinc-400 text-xs uppercase tracking-wider">
                          Username
                        </Label>
                        <Input
                          id="username"
                          type="text"
                          value={loginUsername}
                          onChange={(e) => setLoginUsername(e.target.value)}
                          placeholder="Enter username"
                          className="sentinel-input"
                          required
                          data-testid="login-username"
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="password" className="text-zinc-400 text-xs uppercase tracking-wider">
                          Password
                        </Label>
                        <div className="relative">
                          <Input
                            id="password"
                            type={showPassword ? "text" : "password"}
                            value={loginPassword}
                            onChange={(e) => setLoginPassword(e.target.value)}
                            placeholder="Enter password"
                            className="sentinel-input pr-10"
                            required
                            data-testid="login-password"
                          />
                          <button
                            type="button"
                            onClick={() => setShowPassword(!showPassword)}
                            className="absolute right-3 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-zinc-300"
                          >
                            {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                          </button>
                        </div>
                      </div>
                    </>
                  ) : (
                    <div className="space-y-4">
                      <div className="flex justify-center">
                        <InputOTP
                          maxLength={6}
                          value={totpCode}
                          onChange={(value) => setTotpCode(value)}
                          data-testid="totp-input"
                        >
                          <InputOTPGroup>
                            <InputOTPSlot index={0} className="bg-zinc-950 border-zinc-700" />
                            <InputOTPSlot index={1} className="bg-zinc-950 border-zinc-700" />
                            <InputOTPSlot index={2} className="bg-zinc-950 border-zinc-700" />
                            <InputOTPSlot index={3} className="bg-zinc-950 border-zinc-700" />
                            <InputOTPSlot index={4} className="bg-zinc-950 border-zinc-700" />
                            <InputOTPSlot index={5} className="bg-zinc-950 border-zinc-700" />
                          </InputOTPGroup>
                        </InputOTP>
                      </div>
                      <button
                        type="button"
                        onClick={() => {
                          setRequires2FA(false);
                          setTotpCode("");
                        }}
                        className="text-sm text-cyan-500 hover:text-cyan-400 font-mono"
                      >
                        ← Back to login
                      </button>
                    </div>
                  )}

                  <Button 
                    type="submit" 
                    className="w-full sentinel-btn-primary"
                    disabled={isLoading || (requires2FA && totpCode.length !== 6)}
                    data-testid="login-submit"
                  >
                    {isLoading ? "Authenticating..." : requires2FA ? "Verify Code" : "Login"}
                  </Button>
                </form>
              </CardContent>
            </TabsContent>

            <TabsContent value="register">
              <CardHeader className="pb-4">
                <CardTitle className="text-xl font-bold text-white uppercase tracking-tight">
                  Create Account
                </CardTitle>
                <CardDescription className="text-zinc-500 font-mono text-xs">
                  Register a new operator account
                </CardDescription>
              </CardHeader>
              <CardContent>
                <form onSubmit={handleRegister} className="space-y-4">
                  <div className="space-y-2">
                    <Label htmlFor="reg-username" className="text-zinc-400 text-xs uppercase tracking-wider">
                      Username
                    </Label>
                    <Input
                      id="reg-username"
                      type="text"
                      value={regUsername}
                      onChange={(e) => setRegUsername(e.target.value)}
                      placeholder="Choose a username"
                      className="sentinel-input"
                      required
                      minLength={3}
                      data-testid="register-username"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="reg-email" className="text-zinc-400 text-xs uppercase tracking-wider">
                      Email
                    </Label>
                    <Input
                      id="reg-email"
                      type="email"
                      value={regEmail}
                      onChange={(e) => setRegEmail(e.target.value)}
                      placeholder="Enter email address"
                      className="sentinel-input"
                      required
                      data-testid="register-email"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="reg-password" className="text-zinc-400 text-xs uppercase tracking-wider">
                      Password
                    </Label>
                    <Input
                      id="reg-password"
                      type="password"
                      value={regPassword}
                      onChange={(e) => setRegPassword(e.target.value)}
                      placeholder="Min 12 characters"
                      className="sentinel-input"
                      required
                      minLength={12}
                      data-testid="register-password"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="reg-confirm" className="text-zinc-400 text-xs uppercase tracking-wider">
                      Confirm Password
                    </Label>
                    <Input
                      id="reg-confirm"
                      type="password"
                      value={regConfirmPassword}
                      onChange={(e) => setRegConfirmPassword(e.target.value)}
                      placeholder="Confirm password"
                      className="sentinel-input"
                      required
                      data-testid="register-confirm-password"
                    />
                  </div>

                  <div className="flex items-start gap-2 p-3 bg-yellow-500/10 border border-yellow-500/20 rounded-sm">
                    <AlertTriangle className="w-4 h-4 text-yellow-500 mt-0.5 flex-shrink-0" />
                    <p className="text-xs text-yellow-500/80 font-mono">
                      Password must contain uppercase, lowercase, and numbers
                    </p>
                  </div>

                  <Button 
                    type="submit" 
                    className="w-full sentinel-btn-primary"
                    disabled={isLoading}
                    data-testid="register-submit"
                  >
                    {isLoading ? "Creating Account..." : "Register"}
                  </Button>
                </form>
              </CardContent>
            </TabsContent>
          </Tabs>
        </Card>

        {/* Footer */}
        <p className="text-center text-xs text-zinc-600 font-mono mt-6">
          SENTINEL NOC v1.0.0 • SECURE CONNECTION
        </p>
      </div>
    </div>
  );
}
