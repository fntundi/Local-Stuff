import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { MainLayout } from "@/components/layout/MainLayout";
import { useAuth } from "@/contexts/AuthContext";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { toast } from "sonner";
import { 
  Plus, 
  Camera, 
  Pencil, 
  Trash2, 
  Wifi, 
  WifiOff,
  Eye,
  Video,
  MapPin
} from "lucide-react";

export default function CamerasPage() {
  const { api, user } = useAuth();
  const navigate = useNavigate();
  const [cameras, setCameras] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showAddDialog, setShowAddDialog] = useState(false);
  const [deleteCamera, setDeleteCamera] = useState(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  const [formData, setFormData] = useState({
    name: "",
    ip_address: "",
    port: 80,
    rtsp_port: 554,
    rtsp_path: "/stream1",
    username: "admin",
    password: "",
    location: "",
    ptz_capable: false
  });

  const canManageCameras = ["admin", "security_operator"].includes(user?.role);
  const canDeleteCameras = user?.role === "admin";

  useEffect(() => {
    fetchCameras();
  }, []);

  const fetchCameras = async () => {
    try {
      const response = await api.get("/cameras");
      setCameras(response.data);
    } catch (error) {
      toast.error("Failed to fetch cameras");
    } finally {
      setIsLoading(false);
    }
  };

  const handleAddCamera = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    
    try {
      await api.post("/cameras", formData);
      toast.success("Camera added successfully");
      setShowAddDialog(false);
      resetForm();
      fetchCameras();
    } catch (error) {
      const message = error.response?.data?.detail || "Failed to add camera";
      toast.error(message);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDeleteCamera = async () => {
    if (!deleteCamera) return;
    
    try {
      await api.delete(`/cameras/${deleteCamera.id}`);
      toast.success("Camera deleted successfully");
      setDeleteCamera(null);
      fetchCameras();
    } catch (error) {
      toast.error("Failed to delete camera");
    }
  };

  const resetForm = () => {
    setFormData({
      name: "",
      ip_address: "",
      port: 80,
      rtsp_port: 554,
      rtsp_path: "/stream1",
      username: "admin",
      password: "",
      location: "",
      ptz_capable: false
    });
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

  return (
    <MainLayout>
      <div className="space-y-6 animate-fadeIn" data-testid="cameras-page">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-white uppercase tracking-tight font-['Chivo']">
              Camera Management
            </h1>
            <p className="text-sm text-zinc-500 font-mono">
              {cameras.length} cameras configured
            </p>
          </div>
          {canManageCameras && (
            <Dialog open={showAddDialog} onOpenChange={setShowAddDialog}>
              <DialogTrigger asChild>
                <Button className="sentinel-btn-primary" data-testid="add-camera-btn">
                  <Plus className="w-4 h-4 mr-2" />
                  Add Camera
                </Button>
              </DialogTrigger>
              <DialogContent className="bg-zinc-900 border-zinc-800 max-w-lg">
                <DialogHeader>
                  <DialogTitle className="text-xl font-bold text-white uppercase tracking-tight">
                    Add New Camera
                  </DialogTitle>
                  <DialogDescription className="text-zinc-500 font-mono text-xs">
                    Configure a new ONVIF/RTSP camera
                  </DialogDescription>
                </DialogHeader>
                <form onSubmit={handleAddCamera} className="space-y-4">
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label className="text-zinc-400 text-xs uppercase tracking-wider">
                        Camera Name
                      </Label>
                      <Input
                        value={formData.name}
                        onChange={(e) => setFormData({...formData, name: e.target.value})}
                        placeholder="Main Entrance"
                        className="sentinel-input"
                        required
                        data-testid="camera-name-input"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label className="text-zinc-400 text-xs uppercase tracking-wider">
                        Location
                      </Label>
                      <Input
                        value={formData.location}
                        onChange={(e) => setFormData({...formData, location: e.target.value})}
                        placeholder="Building A, Floor 1"
                        className="sentinel-input"
                        data-testid="camera-location-input"
                      />
                    </div>
                  </div>
                  
                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label className="text-zinc-400 text-xs uppercase tracking-wider">
                        IP Address
                      </Label>
                      <Input
                        value={formData.ip_address}
                        onChange={(e) => setFormData({...formData, ip_address: e.target.value})}
                        placeholder="192.168.1.100"
                        className="sentinel-input"
                        required
                        data-testid="camera-ip-input"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label className="text-zinc-400 text-xs uppercase tracking-wider">
                        HTTP Port
                      </Label>
                      <Input
                        type="number"
                        value={formData.port}
                        onChange={(e) => setFormData({...formData, port: parseInt(e.target.value)})}
                        className="sentinel-input"
                        min={1}
                        max={65535}
                        data-testid="camera-port-input"
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label className="text-zinc-400 text-xs uppercase tracking-wider">
                        RTSP Port
                      </Label>
                      <Input
                        type="number"
                        value={formData.rtsp_port}
                        onChange={(e) => setFormData({...formData, rtsp_port: parseInt(e.target.value)})}
                        className="sentinel-input"
                        min={1}
                        max={65535}
                        data-testid="camera-rtsp-port-input"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label className="text-zinc-400 text-xs uppercase tracking-wider">
                        RTSP Path
                      </Label>
                      <Input
                        value={formData.rtsp_path}
                        onChange={(e) => setFormData({...formData, rtsp_path: e.target.value})}
                        placeholder="/stream1"
                        className="sentinel-input"
                        data-testid="camera-rtsp-path-input"
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label className="text-zinc-400 text-xs uppercase tracking-wider">
                        Username
                      </Label>
                      <Input
                        value={formData.username}
                        onChange={(e) => setFormData({...formData, username: e.target.value})}
                        placeholder="admin"
                        className="sentinel-input"
                        required
                        data-testid="camera-username-input"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label className="text-zinc-400 text-xs uppercase tracking-wider">
                        Password
                      </Label>
                      <Input
                        type="password"
                        value={formData.password}
                        onChange={(e) => setFormData({...formData, password: e.target.value})}
                        placeholder="Camera password"
                        className="sentinel-input"
                        required
                        data-testid="camera-password-input"
                      />
                    </div>
                  </div>

                  <div className="flex items-center justify-between p-3 bg-zinc-800/50 rounded-sm">
                    <div>
                      <Label className="text-zinc-300 text-sm">PTZ Capable</Label>
                      <p className="text-xs text-zinc-500 font-mono">Enable pan-tilt-zoom controls</p>
                    </div>
                    <Switch
                      checked={formData.ptz_capable}
                      onCheckedChange={(checked) => setFormData({...formData, ptz_capable: checked})}
                      data-testid="camera-ptz-switch"
                    />
                  </div>

                  <DialogFooter>
                    <Button
                      type="button"
                      variant="outline"
                      onClick={() => setShowAddDialog(false)}
                      className="bg-zinc-800 border-zinc-700 text-white hover:bg-zinc-700"
                    >
                      Cancel
                    </Button>
                    <Button 
                      type="submit" 
                      className="sentinel-btn-primary"
                      disabled={isSubmitting}
                      data-testid="submit-camera-btn"
                    >
                      {isSubmitting ? "Adding..." : "Add Camera"}
                    </Button>
                  </DialogFooter>
                </form>
              </DialogContent>
            </Dialog>
          )}
        </div>

        {/* Cameras Table */}
        <Card className="sentinel-card">
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow className="border-zinc-800 hover:bg-transparent">
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">Status</TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">Name</TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">IP Address</TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">Location</TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">Features</TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">Last Seen</TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {cameras.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center py-12">
                      <Camera className="w-12 h-12 text-zinc-700 mx-auto mb-3" />
                      <p className="text-zinc-500 font-mono">No cameras configured</p>
                      {canManageCameras && (
                        <p className="text-xs text-zinc-600 font-mono mt-1">
                          Click "Add Camera" to get started
                        </p>
                      )}
                    </TableCell>
                  </TableRow>
                ) : (
                  cameras.map((camera) => (
                    <TableRow 
                      key={camera.id} 
                      className="border-zinc-800 hover:bg-zinc-800/50"
                      data-testid={`camera-row-${camera.id}`}
                    >
                      <TableCell>
                        <div className="flex items-center gap-2">
                          {camera.is_online ? (
                            <>
                              <Wifi className="w-4 h-4 text-emerald-500" />
                              <span className="badge-online">Online</span>
                            </>
                          ) : (
                            <>
                              <WifiOff className="w-4 h-4 text-zinc-500" />
                              <span className="badge-offline">Offline</span>
                            </>
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        <div>
                          <p className="text-white font-medium">{camera.name}</p>
                          {camera.manufacturer && (
                            <p className="text-xs text-zinc-500 font-mono">
                              {camera.manufacturer} {camera.model}
                            </p>
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        <span className="text-zinc-300 font-mono text-sm">
                          {camera.ip_address}:{camera.port}
                        </span>
                      </TableCell>
                      <TableCell>
                        {camera.location ? (
                          <div className="flex items-center gap-1.5 text-zinc-400">
                            <MapPin className="w-3 h-3" />
                            <span className="text-sm">{camera.location}</span>
                          </div>
                        ) : (
                          <span className="text-zinc-600">—</span>
                        )}
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          {camera.ptz_capable && (
                            <span className="text-xs bg-cyan-500/10 text-cyan-500 px-2 py-0.5 rounded-sm border border-cyan-500/20 font-mono">
                              PTZ
                            </span>
                          )}
                          {camera.recording_enabled && (
                            <span className="text-xs bg-red-500/10 text-red-500 px-2 py-0.5 rounded-sm border border-red-500/20 font-mono">
                              REC
                            </span>
                          )}
                          {camera.motion_detection_enabled && (
                            <span className="text-xs bg-yellow-500/10 text-yellow-500 px-2 py-0.5 rounded-sm border border-yellow-500/20 font-mono">
                              MD
                            </span>
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        <span className="text-zinc-500 font-mono text-xs">
                          {camera.last_seen 
                            ? new Date(camera.last_seen).toLocaleString()
                            : "Never"
                          }
                        </span>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center justify-end gap-2">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => navigate(`/cameras/${camera.id}`)}
                            className="text-zinc-400 hover:text-white hover:bg-zinc-800"
                            data-testid={`view-camera-${camera.id}`}
                          >
                            <Eye className="w-4 h-4" />
                          </Button>
                          {canDeleteCameras && (
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => setDeleteCamera(camera)}
                              className="text-zinc-400 hover:text-red-500 hover:bg-red-500/10"
                              data-testid={`delete-camera-${camera.id}`}
                            >
                              <Trash2 className="w-4 h-4" />
                            </Button>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        {/* Delete Confirmation Dialog */}
        <AlertDialog open={!!deleteCamera} onOpenChange={() => setDeleteCamera(null)}>
          <AlertDialogContent className="bg-zinc-900 border-zinc-800">
            <AlertDialogHeader>
              <AlertDialogTitle className="text-white">Delete Camera</AlertDialogTitle>
              <AlertDialogDescription className="text-zinc-400">
                Are you sure you want to delete "{deleteCamera?.name}"? This action cannot be undone.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel className="bg-zinc-800 border-zinc-700 text-white hover:bg-zinc-700">
                Cancel
              </AlertDialogCancel>
              <AlertDialogAction 
                onClick={handleDeleteCamera}
                className="sentinel-btn-destructive"
                data-testid="confirm-delete-camera"
              >
                Delete
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>
    </MainLayout>
  );
}
