import { createContext, useContext, useState, useEffect, useCallback } from "react";
import axios from "axios";

const API_URL = process.env.REACT_APP_BACKEND_URL;

const AuthContext = createContext(null);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [accessToken, setAccessToken] = useState(localStorage.getItem("accessToken"));
  const [refreshToken, setRefreshToken] = useState(localStorage.getItem("refreshToken"));

  const api = axios.create({
    baseURL: `${API_URL}/api`,
    headers: {
      "Content-Type": "application/json",
    },
  });

  // Add auth header to requests
  api.interceptors.request.use((config) => {
    if (accessToken) {
      config.headers.Authorization = `Bearer ${accessToken}`;
    }
    return config;
  });

  // Handle token refresh on 401
  api.interceptors.response.use(
    (response) => response,
    async (error) => {
      const originalRequest = error.config;
      if (error.response?.status === 401 && !originalRequest._retry && refreshToken) {
        originalRequest._retry = true;
        try {
          const response = await axios.post(`${API_URL}/api/auth/refresh`, null, {
            params: { refresh_token: refreshToken }
          });
          const newAccessToken = response.data.access_token;
          setAccessToken(newAccessToken);
          localStorage.setItem("accessToken", newAccessToken);
          originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
          return api(originalRequest);
        } catch (refreshError) {
          logout();
          return Promise.reject(refreshError);
        }
      }
      return Promise.reject(error);
    }
  );

  const fetchUser = useCallback(async () => {
    if (!accessToken) {
      setIsLoading(false);
      return;
    }
    try {
      const response = await api.get("/auth/me");
      setUser(response.data);
    } catch (error) {
      console.error("Failed to fetch user:", error);
      logout();
    } finally {
      setIsLoading(false);
    }
  }, [accessToken]);

  useEffect(() => {
    fetchUser();
  }, [fetchUser]);

  const login = async (username, password, totpCode = null) => {
    const payload = { username, password };
    if (totpCode) {
      payload.totp_code = totpCode;
    }
    
    const response = await api.post("/auth/login", payload);
    
    if (response.data.requires_2fa) {
      return { requires2FA: true };
    }
    
    const { access_token, refresh_token, user: userData } = response.data;
    setAccessToken(access_token);
    setRefreshToken(refresh_token);
    setUser(userData);
    localStorage.setItem("accessToken", access_token);
    localStorage.setItem("refreshToken", refresh_token);
    
    return { success: true };
  };

  const register = async (username, email, password) => {
    const response = await api.post("/auth/register", { username, email, password });
    return response.data;
  };

  const logout = () => {
    setUser(null);
    setAccessToken(null);
    setRefreshToken(null);
    localStorage.removeItem("accessToken");
    localStorage.removeItem("refreshToken");
  };

  const setup2FA = async () => {
    const response = await api.post("/auth/2fa/setup");
    return response.data;
  };

  const verify2FA = async (code) => {
    const response = await api.post("/auth/2fa/verify", { code });
    await fetchUser();
    return response.data;
  };

  const disable2FA = async (code) => {
    const response = await api.post("/auth/2fa/disable", { code });
    await fetchUser();
    return response.data;
  };

  const value = {
    user,
    isLoading,
    isAuthenticated: !!user,
    login,
    register,
    logout,
    setup2FA,
    verify2FA,
    disable2FA,
    api,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};
