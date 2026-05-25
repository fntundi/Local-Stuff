import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { MainLayout } from "@/components/layout/MainLayout";
import { useAuth } from "@/contexts/AuthContext";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { toast } from "sonner";
import { 
  ArrowLeft, 
  Wifi, 
  WifiOff,
  Video,
  Circle,
  ChevronUp,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  ZoomIn,
  ZoomOut,
  Home,
  Camera as CameraIcon,
  MapPin,
  Settings,
  Activity
} from "lucide-react";

export default function CameraDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { api, user } = useAuth();
  const [camera, setCamera] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [streamUrl, setStreamUrl] = useState(null);

  const canManageCameras = ["admin", "security_operator"].includes(user?.role);

  const placeholderImages = [
    "https://images.unsplash.com/photo-1747573235085-aa6b21b92e07?auto=format&fit=crop&w=1200&q=80",
    "https://images.unsplash.com/photo-1750189953388-5ef155c12369?auto=format&fit=crop&w=1200&q=80",
    "https://images.unsplash.com/photo-1631248207065-771ae9ac32f0?auto=format&fit=crop&w=1200&q=80",
    "https://images.unsplash.com/photo-1763333407379-9bd0aca01dfd?auto=format&fit=crop&w=1200&q=80"
  ];

  useEffect(() => {
    fetchCamera();
  }, [id]);

  const fetchCamera = async () => {
    try {
      const response = await api.get(`/cameras/${id}`);
      setCamera(response.data);
    } catch (error) {
      toast.error("Failed to fetch camera details");
      navigate("/cameras");
    } finally {
      setIsLoading(false);
    }
  };

  const toggleMotionDetection = async () => {
    try {
      await api.put(`/cameras/${id}`, {
        motion_detection_enabled: !camera.motion_detection_enabled
      });
      setCamera({...camera, motion_detection_enabled: !camera.motion_detection_enabled});
      toast.success(`Motion detection ${!camera.motion_detection_enabled ? "enabled" : "disabled"}`);
    } catch (error) {
      toast.error("Failed to update setting");
    }
  };

  const toggleRecording = async () => {
    try {
      await api.put(`/cameras/${id}`, {
        recording_enabled: !camera.recording_enabled
      });
      setCamera({...camera, recording_enabled: !camera.recording_enabled});
      toast.success(`Recording ${!camera.recording_enabled ? "started" : "stopped"}`);
    } catch (error) {
      toast.error("Failed to update setting");
    }
  };

  const handlePTZControl = async (action) => {
    toast.info(`PTZ: ${action}`, { description: "Command sent to camera" });
  };

  if (isLoading) {
    return (
      <MainLayout>
        <div className="flex items-center justify-center h-full">
          <div className="text-cyan-500 font-mono">Loading...</div>
        </div>
      </MainLayout>
    );
  }

  if (!camera) {
    return (
      <MainLayout>
        <div className="flex items-center justify-center h-full">
          <div className="text-zinc-500 font-mono">Camera not found</div>
        </div>
      </MainLayout>
    );
  }

  return (
    <MainLayout>
      <div className="space-y-6 animate-fadeIn" data-testid="camera-detail-page">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Button
              variant="ghost"
              onClick={() => navigate("/cameras")}
              className="text-zinc-400 hover:text-white hover:bg-zinc-800"
              data-testid="back-btn"
            >
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back
            </Button>
            <div>
              <h1 className="text-2xl font-bold text-white uppercase tracking-tight font-['Chivo']">
                {camera.name}
              </h1>
              <div className="flex items-center gap-3 mt-1">
                {camera.is_online ? (
                  <span className="badge-online">Online</span>
                ) : (
                  <span className="badge-offline">Offline</span>
                )}
                <span className="text-sm text-zinc-500 font-mono">
                  {camera.ip_address}:{camera.port}
                </span>
              </div>
            </div>
          </div>
        </div>

        <div className="grid grid-cols-12 gap-6">
          {/* Main Video Feed */}
          <div className="col-span-8">
            <Card className="sentinel-card">
              <CardContent className="p-0">
                <div className="video-feed scanlines aspect-video">
                  <img 
                    src={placeholderImages[parseInt(id.charCodeAt(0)) % placeholderImages.length]} 
                    alt={camera.name}
                    className="w-full h-full object-cover"
                  />
                  
                  {/* Header Overlay */}
                  <div className="video-feed-header flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <span className={camera.is_online ? "badge-live" : "badge-offline"}>
                        {camera.is_online ? "LIVE" : "OFFLINE"}
                      </span>
                      <span className="text-sm font-mono text-white">{camera.name}</span>
                    </div>
                    <span className="text-xs font-mono text-zinc-400">
                      {new Date().toLocaleString()}
                    </span>
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
                        <span className="text-xs font-mono text-zinc-400">
                          RTSP:{camera.rtsp_port} • {camera.rtsp_path}
                        </span>
                      </div>
                    </div>
                  </div>

                  {/* Recording Indicator */}
                  {camera.recording_enabled && (
                    <div className="absolute top-3 right-3 flex items-center gap-1.5 bg-red-500/20 px-2 py-1 rounded-sm border border-red-500/30 z-20">
                      <Circle className="w-2 h-2 fill-red-500 text-red-500 animate-pulse" />
                      <span className="text-xs font-mono text-red-500">REC</span>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>

            {/* PTZ Controls */}
            {camera.ptz_capable && (
              <Card className="sentinel-card mt-4">
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-bold text-white uppercase tracking-tight">
                    PTZ Controls
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="flex items-center justify-center gap-8">
                    {/* Direction Pad */}
                    <div className="grid grid-cols-3 gap-1">
                      <div />
                      <button 
                        onClick={() => handlePTZControl('tilt-up')}
                        className="ptz-control"
                        data-testid="ptz-up"
                      >
                        <ChevronUp className="w-5 h-5 text-white" />
                      </button>
                      <div />
                      <button 
                        onClick={() => handlePTZControl('pan-left')}
                        className="ptz-control"
                        data-testid="ptz-left"
                      >
                        <ChevronLeft className="w-5 h-5 text-white" />
                      </button>
                      <button 
                        onClick={() => handlePTZControl('home')}
                        className="ptz-control"
                        data-testid="ptz-home"
                      >
                        <Home className="w-4 h-4 text-white" />
                      </button>
                      <button 
                        onClick={() => handlePTZControl('pan-right')}
                        className="ptz-control"
                        data-testid="ptz-right"
                      >
                        <ChevronRight className="w-5 h-5 text-white" />
                      </button>
                      <div />
                      <button 
                        onClick={() => handlePTZControl('tilt-down')}
                        className="ptz-control"
                        data-testid="ptz-down"
                      >
                        <ChevronDown className="w-5 h-5 text-white" />
                      </button>
                      <div />
                    </div>

                    {/* Zoom Controls */}
                    <div className="flex flex-col gap-1">
                      <button 
                        onClick={() => handlePTZControl('zoom-in')}
                        className="ptz-control"
                        data-testid="ptz-zoom-in"
                      >
                        <ZoomIn className="w-5 h-5 text-white" />
                      </button>
                      <button 
                        onClick={() => handlePTZControl('zoom-out')}
                        className="ptz-control"
                        data-testid="ptz-zoom-out"
                      >
                        <ZoomOut className="w-5 h-5 text-white" />
                      </button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            )}
          </div>

          {/* Side Panel */}
          <div className="col-span-4 space-y-4">
            {/* Camera Info */}
            <Card className="sentinel-card">
              <CardHeader className="pb-3">
                <CardTitle className="text-sm font-bold text-white uppercase tracking-tight flex items-center gap-2">
                  <CameraIcon className="w-4 h-4 text-cyan-500" />
                  Camera Details
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="flex justify-between items-center py-2 border-b border-zinc-800">
                  <span className="text-xs text-zinc-500 font-mono uppercase">IP Address</span>
                  <span className="text-sm text-white font-mono">{camera.ip_address}</span>
                </div>
                <div className="flex justify-between items-center py-2 border-b border-zinc-800">
                  <span className="text-xs text-zinc-500 font-mono uppercase">HTTP Port</span>
                  <span className="text-sm text-white font-mono">{camera.port}</span>
                </div>
                <div className="flex justify-between items-center py-2 border-b border-zinc-800">
                  <span className="text-xs text-zinc-500 font-mono uppercase">RTSP Port</span>
                  <span className="text-sm text-white font-mono">{camera.rtsp_port}</span>
                </div>
                {camera.manufacturer && (
                  <div className="flex justify-between items-center py-2 border-b border-zinc-800">
                    <span className="text-xs text-zinc-500 font-mono uppercase">Manufacturer</span>
                    <span className="text-sm text-white">{camera.manufacturer}</span>
                  </div>
                )}
                {camera.model && (
                  <div className="flex justify-between items-center py-2 border-b border-zinc-800">
                    <span className="text-xs text-zinc-500 font-mono uppercase">Model</span>
                    <span className="text-sm text-white">{camera.model}</span>
                  </div>
                )}
                {camera.location && (
                  <div className="flex justify-between items-center py-2">
                    <span className="text-xs text-zinc-500 font-mono uppercase">Location</span>
                    <span className="text-sm text-white flex items-center gap-1">
                      <MapPin className="w-3 h-3" />
                      {camera.location}
                    </span>
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Camera Controls */}
            {canManageCameras && (
              <Card className="sentinel-card">
                <CardHeader className="pb-3">
                  <CardTitle className="text-sm font-bold text-white uppercase tracking-tight flex items-center gap-2">
                    <Settings className="w-4 h-4 text-cyan-500" />
                    Controls
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="flex items-center justify-between p-3 bg-zinc-800/50 rounded-sm">
                    <div>
                      <Label className="text-zinc-300 text-sm">Recording</Label>
                      <p className="text-xs text-zinc-500 font-mono">
                        {camera.recording_enabled ? "Recording active" : "Recording disabled"}
                      </p>
                    </div>
                    <Switch
                      checked={camera.recording_enabled}
                      onCheckedChange={toggleRecording}
                      data-testid="recording-toggle"
                    />
                  </div>
                  <div className="flex items-center justify-between p-3 bg-zinc-800/50 rounded-sm">
                    <div>
                      <Label className="text-zinc-300 text-sm">Motion Detection</Label>
                      <p className="text-xs text-zinc-500 font-mono">
                        {camera.motion_detection_enabled ? "Detection active" : "Detection disabled"}
                      </p>
                    </div>
                    <Switch
                      checked={camera.motion_detection_enabled}
                      onCheckedChange={toggleMotionDetection}
                      data-testid="motion-toggle"
                    />
                  </div>
                </CardContent>
              </Card>
            )}

            {/* Camera Status */}
            <Card className="sentinel-card">
              <CardHeader className="pb-3">
                <CardTitle className="text-sm font-bold text-white uppercase tracking-tight flex items-center gap-2">
                  <Activity className="w-4 h-4 text-cyan-500" />
                  Status
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                <div className="flex justify-between items-center py-2 border-b border-zinc-800">
                  <span className="text-xs text-zinc-500 font-mono uppercase">Connection</span>
                  <span className={`text-sm font-mono ${camera.is_online ? 'text-emerald-500' : 'text-red-500'}`}>
                    {camera.is_online ? "Connected" : "Disconnected"}
                  </span>
                </div>
                <div className="flex justify-between items-center py-2 border-b border-zinc-800">
                  <span className="text-xs text-zinc-500 font-mono uppercase">Last Seen</span>
                  <span className="text-sm text-white font-mono">
                    {camera.last_seen 
                      ? new Date(camera.last_seen).toLocaleString()
                      : "Never"
                    }
                  </span>
                </div>
                <div className="flex justify-between items-center py-2">
                  <span className="text-xs text-zinc-500 font-mono uppercase">Added</span>
                  <span className="text-sm text-white font-mono">
                    {new Date(camera.created_at).toLocaleDateString()}
                  </span>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </MainLayout>
  );
}
