import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { Toaster } from "@/components/ui/sonner";
import { AuthProvider, useAuth } from "@/contexts/AuthContext";
import LoginPage from "@/pages/LoginPage";
import DashboardPage from "@/pages/DashboardPage";
import CamerasPage from "@/pages/CamerasPage";
import CameraDetailPage from "@/pages/CameraDetailPage";
import EventsPage from "@/pages/EventsPage";
import UsersPage from "@/pages/UsersPage";
import SettingsPage from "@/pages/SettingsPage";
import AuditLogsPage from "@/pages/AuditLogsPage";
import "@/App.css";

const ProtectedRoute = ({ children, allowedRoles }) => {
  const { user, isAuthenticated, isLoading } = useAuth();
  
  if (isLoading) {
    return (
      <div className="min-h-screen bg-[#09090b] flex items-center justify-center">
        <div className="text-cyan-500 font-mono">Loading...</div>
      </div>
    );
  }
  
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  
  if (allowedRoles && !allowedRoles.includes(user?.role)) {
    return <Navigate to="/dashboard" replace />;
  }
  
  return children;
};

function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/dashboard" element={
        <ProtectedRoute>
          <DashboardPage />
        </ProtectedRoute>
      } />
      <Route path="/cameras" element={
        <ProtectedRoute>
          <CamerasPage />
        </ProtectedRoute>
      } />
      <Route path="/cameras/:id" element={
        <ProtectedRoute>
          <CameraDetailPage />
        </ProtectedRoute>
      } />
      <Route path="/events" element={
        <ProtectedRoute>
          <EventsPage />
        </ProtectedRoute>
      } />
      <Route path="/users" element={
        <ProtectedRoute allowedRoles={["admin"]}>
          <UsersPage />
        </ProtectedRoute>
      } />
      <Route path="/settings" element={
        <ProtectedRoute allowedRoles={["admin"]}>
          <SettingsPage />
        </ProtectedRoute>
      } />
      <Route path="/audit-logs" element={
        <ProtectedRoute allowedRoles={["admin"]}>
          <AuditLogsPage />
        </ProtectedRoute>
      } />
      <Route path="/" element={<Navigate to="/dashboard" replace />} />
      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  );
}

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <AppRoutes />
        <Toaster 
          position="top-right" 
          toastOptions={{
            style: {
              background: '#18181b',
              border: '1px solid #27272a',
              color: '#fafafa',
            },
          }}
        />
      </AuthProvider>
    </BrowserRouter>
  );
}

export default App;
