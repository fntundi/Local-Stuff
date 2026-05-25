import { useState, useEffect } from "react";
import { MainLayout } from "@/components/layout/MainLayout";
import { useAuth } from "@/contexts/AuthContext";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Slider } from "@/components/ui/slider";
import { toast } from "sonner";
import { 
  Settings, 
  HardDrive, 
  Bell, 
  Mail, 
  Save,
  AlertTriangle
} from "lucide-react";

export default function SettingsPage() {
  const { api } = useAuth();
  const [settings, setSettings] = useState({
    storage_path: "/recordings",
    retention_days: 30,
    motion_sensitivity: 50,
    alarm_notification_enabled: true,
    email_notifications: false,
    smtp_server: "",
    smtp_port: 587
  });
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    fetchSettings();
  }, []);

  const fetchSettings = async () => {
    try {
      const response = await api.get("/settings");
      setSettings(response.data);
    } catch (error) {
      toast.error("Failed to fetch settings");
    } finally {
      setIsLoading(false);
    }
  };

  const saveSettings = async () => {
    setIsSaving(true);
    try {
      await api.put("/settings", settings);
      toast.success("Settings saved successfully");
    } catch (error) {
      toast.error("Failed to save settings");
    } finally {
      setIsSaving(false);
    }
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
      <div className="space-y-6 animate-fadeIn" data-testid="settings-page">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-white uppercase tracking-tight font-['Chivo']">
              System Settings
            </h1>
            <p className="text-sm text-zinc-500 font-mono">Configure system parameters</p>
          </div>
          <Button 
            onClick={saveSettings} 
            className="sentinel-btn-primary"
            disabled={isSaving}
            data-testid="save-settings-btn"
          >
            <Save className="w-4 h-4 mr-2" />
            {isSaving ? "Saving..." : "Save Settings"}
          </Button>
        </div>

        <div className="grid grid-cols-2 gap-6">
          {/* Storage Settings */}
          <Card className="sentinel-card">
            <CardHeader className="pb-4">
              <CardTitle className="text-lg font-bold text-white uppercase tracking-tight flex items-center gap-2">
                <HardDrive className="w-5 h-5 text-cyan-500" />
                Storage
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="space-y-2">
                <Label className="text-zinc-400 text-xs uppercase tracking-wider">
                  Storage Path
                </Label>
                <Input
                  value={settings.storage_path}
                  onChange={(e) => setSettings({...settings, storage_path: e.target.value})}
                  placeholder="/recordings"
                  className="sentinel-input"
                  data-testid="storage-path-input"
                />
                <p className="text-xs text-zinc-600 font-mono">
                  Path where recordings will be stored
                </p>
              </div>

              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <Label className="text-zinc-400 text-xs uppercase tracking-wider">
                    Retention Period
                  </Label>
                  <span className="text-sm text-cyan-500 font-mono">
                    {settings.retention_days} days
                  </span>
                </div>
                <Slider
                  value={[settings.retention_days]}
                  onValueChange={([value]) => setSettings({...settings, retention_days: value})}
                  min={7}
                  max={90}
                  step={1}
                  className="w-full"
                  data-testid="retention-slider"
                />
                <div className="flex justify-between text-xs text-zinc-600 font-mono">
                  <span>7 days</span>
                  <span>90 days</span>
                </div>
              </div>

              <div className="p-3 bg-yellow-500/10 border border-yellow-500/20 rounded-sm flex items-start gap-2">
                <AlertTriangle className="w-4 h-4 text-yellow-500 mt-0.5 flex-shrink-0" />
                <div>
                  <p className="text-xs text-yellow-500 font-medium">Security Notice</p>
                  <p className="text-xs text-yellow-500/70 font-mono mt-1">
                    NAS storage is unencrypted. Ensure NAS-level encryption and access controls are enabled.
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Motion Detection Settings */}
          <Card className="sentinel-card">
            <CardHeader className="pb-4">
              <CardTitle className="text-lg font-bold text-white uppercase tracking-tight flex items-center gap-2">
                <Bell className="w-5 h-5 text-cyan-500" />
                Motion Detection
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <Label className="text-zinc-400 text-xs uppercase tracking-wider">
                    Sensitivity
                  </Label>
                  <span className="text-sm text-cyan-500 font-mono">
                    {settings.motion_sensitivity}%
                  </span>
                </div>
                <Slider
                  value={[settings.motion_sensitivity]}
                  onValueChange={([value]) => setSettings({...settings, motion_sensitivity: value})}
                  min={10}
                  max={100}
                  step={5}
                  className="w-full"
                  data-testid="sensitivity-slider"
                />
                <div className="flex justify-between text-xs text-zinc-600 font-mono">
                  <span>Low</span>
                  <span>High</span>
                </div>
              </div>

              <div className="flex items-center justify-between p-3 bg-zinc-800/50 rounded-sm">
                <div>
                  <Label className="text-zinc-300 text-sm">Alarm Notifications</Label>
                  <p className="text-xs text-zinc-500 font-mono">
                    Send alerts on motion detection
                  </p>
                </div>
                <Switch
                  checked={settings.alarm_notification_enabled}
                  onCheckedChange={(checked) => setSettings({...settings, alarm_notification_enabled: checked})}
                  data-testid="alarm-notifications-switch"
                />
              </div>
            </CardContent>
          </Card>

          {/* Email Notifications */}
          <Card className="sentinel-card col-span-2">
            <CardHeader className="pb-4">
              <CardTitle className="text-lg font-bold text-white uppercase tracking-tight flex items-center gap-2">
                <Mail className="w-5 h-5 text-cyan-500" />
                Email Notifications
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="flex items-center justify-between p-3 bg-zinc-800/50 rounded-sm">
                <div>
                  <Label className="text-zinc-300 text-sm">Enable Email Notifications</Label>
                  <p className="text-xs text-zinc-500 font-mono">
                    Send event notifications via email
                  </p>
                </div>
                <Switch
                  checked={settings.email_notifications}
                  onCheckedChange={(checked) => setSettings({...settings, email_notifications: checked})}
                  data-testid="email-notifications-switch"
                />
              </div>

              {settings.email_notifications && (
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label className="text-zinc-400 text-xs uppercase tracking-wider">
                      SMTP Server
                    </Label>
                    <Input
                      value={settings.smtp_server}
                      onChange={(e) => setSettings({...settings, smtp_server: e.target.value})}
                      placeholder="smtp.example.com"
                      className="sentinel-input"
                      data-testid="smtp-server-input"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label className="text-zinc-400 text-xs uppercase tracking-wider">
                      SMTP Port
                    </Label>
                    <Input
                      type="number"
                      value={settings.smtp_port}
                      onChange={(e) => setSettings({...settings, smtp_port: parseInt(e.target.value)})}
                      placeholder="587"
                      className="sentinel-input"
                      data-testid="smtp-port-input"
                    />
                  </div>
                </div>
              )}

              <div className="p-3 bg-zinc-800/50 border border-zinc-700 rounded-sm">
                <p className="text-xs text-zinc-500 font-mono">
                  Note: Email notifications will send alerts outside the LAN. Ensure SMTP credentials are configured securely.
                </p>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </MainLayout>
  );
}
