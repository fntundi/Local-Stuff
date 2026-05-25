import { NavLink, useNavigate } from "react-router-dom";
import { useAuth } from "@/contexts/AuthContext";
import { 
  LayoutGrid, 
  Camera, 
  Bell, 
  Users, 
  Settings, 
  FileText, 
  LogOut,
  Shield,
  Activity
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";

export const Sidebar = () => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  const navItems = [
    { to: "/dashboard", icon: LayoutGrid, label: "Dashboard", roles: ["admin", "security_operator", "viewer"] },
    { to: "/cameras", icon: Camera, label: "Cameras", roles: ["admin", "security_operator", "viewer"] },
    { to: "/events", icon: Bell, label: "Events", roles: ["admin", "security_operator", "viewer"] },
    { to: "/users", icon: Users, label: "Users", roles: ["admin"] },
    { to: "/settings", icon: Settings, label: "Settings", roles: ["admin"] },
    { to: "/audit-logs", icon: FileText, label: "Audit Logs", roles: ["admin"] },
  ];

  const filteredNavItems = navItems.filter(item => item.roles.includes(user?.role));

  return (
    <aside className="sidebar flex flex-col" data-testid="sidebar">
      {/* Logo */}
      <div className="p-6 border-b border-zinc-800">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-cyan-500/10 rounded-sm flex items-center justify-center border border-cyan-500/30">
            <Shield className="w-6 h-6 text-cyan-500" />
          </div>
          <div>
            <h1 className="text-lg font-bold text-white tracking-tight uppercase font-['Chivo']">
              SENTINEL
            </h1>
            <p className="text-xs text-zinc-500 font-mono uppercase tracking-wider">NOC System</p>
          </div>
        </div>
      </div>

      {/* System Status */}
      <div className="p-4 border-b border-zinc-800">
        <div className="flex items-center gap-2 text-xs text-zinc-500 font-mono uppercase">
          <Activity className="w-3 h-3 text-emerald-500" />
          <span>System Online</span>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-4 space-y-1">
        {filteredNavItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            data-testid={`nav-${item.label.toLowerCase().replace(' ', '-')}`}
            className={({ isActive }) =>
              `flex items-center gap-3 px-4 py-3 rounded-sm text-sm font-medium uppercase tracking-wide transition-colors ${
                isActive
                  ? "bg-cyan-500/10 text-cyan-500 border-l-2 border-cyan-500"
                  : "text-zinc-400 hover:bg-zinc-800/50 hover:text-white border-l-2 border-transparent"
              }`
            }
          >
            <item.icon className="w-5 h-5" />
            <span>{item.label}</span>
          </NavLink>
        ))}
      </nav>

      <Separator className="bg-zinc-800" />

      {/* User Info */}
      <div className="p-4">
        <div className="sentinel-card">
          <div className="flex items-center gap-3 mb-3">
            <div className="w-8 h-8 bg-zinc-800 rounded-sm flex items-center justify-center">
              <Users className="w-4 h-4 text-zinc-400" />
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-white truncate">{user?.username}</p>
              <p className="text-xs text-zinc-500 font-mono uppercase">{user?.role?.replace('_', ' ')}</p>
            </div>
          </div>
          <div className="flex items-center gap-2 text-xs text-zinc-500 mb-3">
            <span className={`w-2 h-2 rounded-full ${user?.totp_enabled ? 'bg-emerald-500' : 'bg-yellow-500'}`} />
            <span className="font-mono">{user?.totp_enabled ? '2FA ENABLED' : '2FA DISABLED'}</span>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={handleLogout}
            className="w-full bg-zinc-800 border-zinc-700 hover:bg-zinc-700 text-zinc-300 font-medium uppercase tracking-wider"
            data-testid="logout-btn"
          >
            <LogOut className="w-4 h-4 mr-2" />
            Logout
          </Button>
        </div>
      </div>
    </aside>
  );
};
