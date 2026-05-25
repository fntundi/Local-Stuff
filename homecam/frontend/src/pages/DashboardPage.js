import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { MainLayout } from "@/components/layout/MainLayout";
import { useAuth } from "@/contexts/AuthContext";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { 
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { toast } from "sonner";
import { 
  Camera, 
  AlertTriangle, 
  Activity, 
  Bell,
  Grid2X2,
  Grid3X3,
  Square,
  Maximize2,
  Volume2,
  Circle,
  Wifi,
  WifiOff,
  Shield,
  ShieldAlert,
  Home,
  ShieldCheck,
  Siren
} from "lucide-react";

export default function DashboardPage() {
  const { api, user } = useAuth();
  const navigate = useNavigate();
  const [stats, setStats] = useState(null);
  const [cameras, setCameras] = useState([]);
  const [events, setEvents] = useState([]);
  const [gridLayout, setGridLayout] = useState("2x2");
  const [selectedCamera, setSelectedCamera] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [systemMode, setSystemMode] = useState("home");
  const [isChangingMode, setIsChangingMode] = useState(false);

  useEffect(() => {
    fetchDashboardData();
    const interval = setInterval(fetchDashboardData, 30000);
    return () => clearInterval(interval);
  }, []);

  const fetchDashboardData = async () => {
    try {
      const [statsRes, camerasRes, eventsRes] = await Promise.all([
        api.get("/dashboard/stats"),
        api.get("/cameras"),
        api.get("/events", { params: { limit: 10 } })
      ]);
      setStats(statsRes.data);
      setCameras(camerasRes.data);
      setEvents(eventsRes.data);
      // Set system mode from stats
      if (statsRes.data?.system_mode) {
        setSystemMode(statsRes.data.system_mode);
      }
    } catch (error) {
      toast.error("Failed to fetch dashboard data");
    } finally {
      setIsLoading(false);
    }
  };

  const handleModeToggle = async () => {
    const newMode = systemMode === "home" ? "away" : "home";
    setIsChangingMode(true);
    try {
      await api.put("/system/mode", { mode: newMode });
      setSystemMode(newMode);
      toast.success(`System mode changed to ${newMode.toUpperCase()}`, {
        description: newMode === "away" 
          ? "Alarms will trigger on motion detection" 
          : "Alarms disabled for motion events",
        icon: newMode === "away" ? <ShieldAlert className="w-4 h-4" /> : <Home className="w-4 h-4" />
      });
      fetchDashboardData();
    } catch (error) {
      toast.error("Failed to change system mode");
    } finally {
      setIsChangingMode(false);
    }
  };

  const getGridClass = () => {
    switch (gridLayout) {
      case "1x1": return "camera-grid-1x1";
      case "2x2": return "camera-grid-2x2";
      case "3x3": return "camera-grid-3x3";
      case "4x4": return "camera-grid-4x4";
      default: return "camera-grid-2x2";
    }
  };

  const getMaxCameras = () => {
    switch (gridLayout) {
      case "1x1": return 1;
      case "2x2": return 4;
      case "3x3": return 9;
      case "4x4": return 16;
      default: return 4;
    }
  };

  const displayedCameras = cameras.slice(0, getMaxCameras());

  const StatCard = ({ title, value, icon: Icon, color, subValue }) => (
    <Card className="sentinel-card" data-testid={`stat-${title.toLowerCase().replace(' ', '-')}`}>
      <CardContent className="p-4">
        <div className="flex items-start justify-between">
          <div>
            <p className="text-xs text-zinc-500 font-mono uppercase tracking-wider">{title}</p>
            <p className="text-3xl font-bold text-white mt-1 font-['Chivo']">{value}</p>
            {subValue && (
              <p className="text-xs text-zinc-500 font-mono mt-1">{subValue}</p>
            )}
          </div>
          <div className={`p-3 rounded-sm bg-${color}-500/10 border border-${color}-500/20`}>
            <Icon className={`w-6 h-6 text-${color}-500`} />
          </div>
        </div>
      </CardContent>
    </Card>
  );

  const CameraFeed = ({ camera, index }) => {
    const placeholderImages = [
      "https://images.unsplash.com/photo-1747573235085-aa6b21b92e07?auto=format&fit=crop&w=800&q=80",
      "https://images.unsplash.com/photo-1750189953388-5ef155c12369?auto=format&fit=crop&w=800&q=80",
      "https://images.unsplash.com/photo-1631248207065-771ae9ac32f0?auto=format&fit=crop&w=800&q=80",
      "https://images.unsplash.com/photo-1763333407379-9bd0aca01dfd?auto=format&fit=crop&w=800&q=80"
    ];

    return (
      <div 
        className="video-feed scanlines group cursor-pointer"
        onClick={() => navigate(`/cameras/${camera.id}`)}
        data-testid={`camera-feed-${camera.id}`}
      >
        {/* Video/Image Placeholder */}
        <img 
          src={placeholderImages[index % placeholderImages.length]} 
          alt={camera.name}
          className="w-full h-full object-cover"
        />
        
        {/* Header Overlay */}
        <div className="video-feed-header flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className={camera.is_online ? "badge-live" : "badge-offline"}>
              {camera.is_online ? "LIVE" : "OFFLINE"}
            </span>
            <span className="text-xs font-mono text-white">{camera.name}</span>
          </div>
          <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
            <button className="p-1.5 bg-zinc-800/80 rounded-sm hover:bg-zinc-700">
              <Volume2 className="w-4 h-4 text-white" />
            </button>
            <button className="p-1.5 bg-zinc-800/80 rounded-sm hover:bg-zinc-700">
              <Maximize2 className="w-4 h-4 text-white" />
            </button>
          </div>
        </div>

        {/* Footer Overlay */}
        <div className="video-feed-footer">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              {camera.is_online ? (
                <Wifi className="w-3 h-3 text-emerald-500" />
              ) : (
                <WifiOff className="w-3 h-3 text-red-500" />
              )}
              <span className="text-xs font-mono text-zinc-400">{camera.ip_address}</span>
            </div>
            <span className="text-xs font-mono text-zinc-500">
              {new Date().toLocaleTimeString()}
            </span>
          </div>
        </div>

        {/* Recording Indicator */}
        {camera.recording_enabled && (
          <div className="absolute top-2 right-2 flex items-center gap-1.5 bg-red-500/20 px-2 py-1 rounded-sm border border-red-500/30 z-20">
            <Circle className="w-2 h-2 fill-red-500 text-red-500 animate-pulse" />
            <span className="text-xs font-mono text-red-500">REC</span>
          </div>
        )}
      </div>
    );
  };

  const EmptyFeed = ({ index }) => (
    <div className="video-feed flex items-center justify-center" data-testid={`empty-feed-${index}`}>
      <div className="text-center">
        <Camera className="w-12 h-12 text-zinc-700 mx-auto mb-2" />
        <p className="text-xs text-zinc-600 font-mono uppercase">No Camera</p>
      </div>
    </div>
  );

  if (isLoading) {
    return (
      <MainLayout>
        <div className="flex items-center justify-center h-full">
          <div className="text-cyan-500 font-mono">Loading...</div>
        </div>
      </MainLayout>
    );
  }

  return (
    <MainLayout>
      <div className="space-y-6 animate-fadeIn" data-testid="dashboard-page">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-white uppercase tracking-tight font-['Chivo']">
              Operations Dashboard
            </h1>
            <p className="text-sm text-zinc-500 font-mono">Real-time surveillance overview</p>
          </div>
          <div className="flex items-center gap-4">
            {/* System Mode Toggle */}
            <div 
              className={`flex items-center gap-3 px-4 py-2 rounded-sm border transition-all ${
                systemMode === "away" 
                  ? "bg-red-500/10 border-red-500/30" 
                  : "bg-emerald-500/10 border-emerald-500/30"
              }`}
              data-testid="system-mode-toggle"
            >
              {systemMode === "away" ? (
                <ShieldAlert className="w-5 h-5 text-red-500" />
              ) : (
                <Home className="w-5 h-5 text-emerald-500" />
              )}
              <div className="flex flex-col">
                <span className={`text-xs font-mono uppercase ${
                  systemMode === "away" ? "text-red-500" : "text-emerald-500"
                }`}>
                  {systemMode === "away" ? "AWAY MODE" : "HOME MODE"}
                </span>
                <span className="text-xs text-zinc-500">
                  {systemMode === "away" ? "Alarms Active" : "Alarms Off"}
                </span>
              </div>
              {(user?.role === "admin" || user?.role === "security_operator") && (
                <Switch
                  checked={systemMode === "away"}
                  onCheckedChange={handleModeToggle}
                  disabled={isChangingMode}
                  className="ml-2"
                />
              )}
            </div>

            <Select value={gridLayout} onValueChange={setGridLayout}>
              <SelectTrigger className="w-32 bg-zinc-900 border-zinc-800 text-white" data-testid="grid-layout-select">
                <SelectValue placeholder="Grid" />
              </SelectTrigger>
              <SelectContent className="bg-zinc-900 border-zinc-800">
                <SelectItem value="1x1" className="text-white hover:bg-zinc-800">
                  <div className="flex items-center gap-2">
                    <Square className="w-4 h-4" /> 1x1
                  </div>
                </SelectItem>
                <SelectItem value="2x2" className="text-white hover:bg-zinc-800">
                  <div className="flex items-center gap-2">
                    <Grid2X2 className="w-4 h-4" /> 2x2
                  </div>
                </SelectItem>
                <SelectItem value="3x3" className="text-white hover:bg-zinc-800">
                  <div className="flex items-center gap-2">
                    <Grid3X3 className="w-4 h-4" /> 3x3
                  </div>
                </SelectItem>
                <SelectItem value="4x4" className="text-white hover:bg-zinc-800">
                  <div className="flex items-center gap-2">
                    <Grid3X3 className="w-4 h-4" /> 4x4
                  </div>
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        {/* Stats Row */}
        <div className="grid grid-cols-5 gap-4">
          <StatCard 
            title="System Mode" 
            value={systemMode === "away" ? "AWAY" : "HOME"} 
            icon={systemMode === "away" ? ShieldAlert : Shield} 
            color={systemMode === "away" ? "red" : "emerald"}
            subValue={`${stats?.alarm_capable_cameras || 0} alarm-capable`}
          />
          <StatCard 
            title="Total Cameras" 
            value={stats?.total_cameras || 0} 
            icon={Camera} 
            color="cyan"
            subValue={`${stats?.online_cameras || 0} online`}
          />
          <StatCard 
            title="Online" 
            value={stats?.online_cameras || 0} 
            icon={Activity} 
            color="emerald"
            subValue={`${stats?.offline_cameras || 0} offline`}
          />
          <StatCard 
            title="Events" 
            value={stats?.unacknowledged_events || 0} 
            icon={Bell} 
            color="yellow"
            subValue="Unacknowledged"
          />
          <StatCard 
            title="Critical" 
            value={stats?.critical_events || 0} 
            icon={AlertTriangle} 
            color="red"
            subValue="Requires attention"
          />
        </div>

        {/* Main Content Grid */}
        <div className="grid grid-cols-12 gap-6">
          {/* Camera Grid */}
          <div className="col-span-9">
            <Card className="sentinel-card">
              <CardHeader className="pb-4">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-lg font-bold text-white uppercase tracking-tight">
                    Live Feeds
                  </CardTitle>
                  <span className="text-xs font-mono text-zinc-500">
                    {displayedCameras.length} / {cameras.length} cameras
                  </span>
                </div>
              </CardHeader>
              <CardContent>
                <div className={getGridClass()}>
                  {Array.from({ length: getMaxCameras() }).map((_, index) => 
                    displayedCameras[index] ? (
                      <CameraFeed key={displayedCameras[index].id} camera={displayedCameras[index]} index={index} />
                    ) : (
                      <EmptyFeed key={`empty-${index}`} index={index} />
                    )
                  )}
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Events Panel */}
          <div className="col-span-3">
            <Card className="sentinel-card h-full">
              <CardHeader className="pb-4">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-lg font-bold text-white uppercase tracking-tight">
                    Recent Events
                  </CardTitle>
                  <Button 
                    variant="ghost" 
                    size="sm" 
                    onClick={() => navigate("/events")}
                    className="text-cyan-500 hover:text-cyan-400 text-xs font-mono"
                    data-testid="view-all-events-btn"
                  >
                    View All
                  </Button>
                </div>
              </CardHeader>
              <CardContent className="space-y-3">
                {events.length === 0 ? (
                  <div className="text-center py-8">
                    <Bell className="w-8 h-8 text-zinc-700 mx-auto mb-2" />
                    <p className="text-xs text-zinc-600 font-mono">No events</p>
                  </div>
                ) : (
                  events.map((event) => (
                    <div 
                      key={event.id}
                      className={`p-3 rounded-sm border ${
                        event.severity === 'critical' 
                          ? 'bg-red-500/10 border-red-500/20' 
                          : event.severity === 'warning'
                          ? 'bg-yellow-500/10 border-yellow-500/20'
                          : 'bg-zinc-800/50 border-zinc-700'
                      }`}
                      data-testid={`event-${event.id}`}
                    >
                      <div className="flex items-start gap-2">
                        <AlertTriangle className={`w-4 h-4 mt-0.5 flex-shrink-0 ${
                          event.severity === 'critical' ? 'text-red-500' :
                          event.severity === 'warning' ? 'text-yellow-500' : 'text-zinc-400'
                        }`} />
                        <div className="flex-1 min-w-0">
                          <p className="text-sm text-white truncate">{event.message}</p>
                          <div className="flex items-center gap-2 mt-1">
                            <span className="text-xs font-mono text-zinc-500">{event.camera_name}</span>
                            <span className="text-xs font-mono text-zinc-600">
                              {new Date(event.created_at).toLocaleTimeString()}
                            </span>
                          </div>
                        </div>
                      </div>
                    </div>
                  ))
                )}
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </MainLayout>
  );
}
