import { useState, useEffect } from "react";
import { MainLayout } from "@/components/layout/MainLayout";
import { useAuth } from "@/contexts/AuthContext";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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
  FileText, 
  RefreshCw, 
  Filter,
  User,
  Clock,
  Globe
} from "lucide-react";

export default function AuditLogsPage() {
  const { api } = useAuth();
  const [logs, setLogs] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [filters, setFilters] = useState({
    action: "all",
    resource_type: "all"
  });

  useEffect(() => {
    fetchLogs();
  }, [filters]);

  const fetchLogs = async () => {
    try {
      const params = {};
      if (filters.action !== "all") params.action = filters.action;
      if (filters.resource_type !== "all") params.resource_type = filters.resource_type;
      
      const response = await api.get("/audit-logs", { params });
      setLogs(response.data);
    } catch (error) {
      toast.error("Failed to fetch audit logs");
    } finally {
      setIsLoading(false);
    }
  };

  const getActionBadge = (action) => {
    const actionColors = {
      login_success: "bg-emerald-500/10 text-emerald-500 border-emerald-500/20",
      login_failed: "bg-red-500/10 text-red-500 border-red-500/20",
      user_registered: "bg-blue-500/10 text-blue-500 border-blue-500/20",
      camera_created: "bg-cyan-500/10 text-cyan-500 border-cyan-500/20",
      camera_deleted: "bg-red-500/10 text-red-500 border-red-500/20",
      camera_updated: "bg-yellow-500/10 text-yellow-500 border-yellow-500/20",
      settings_updated: "bg-purple-500/10 text-purple-500 border-purple-500/20",
      "2fa_enabled": "bg-emerald-500/10 text-emerald-500 border-emerald-500/20",
      "2fa_disabled": "bg-yellow-500/10 text-yellow-500 border-yellow-500/20",
    };
    
    const colorClass = actionColors[action] || "bg-zinc-500/10 text-zinc-500 border-zinc-500/20";
    
    return (
      <span className={`text-xs px-2 py-0.5 rounded-sm font-mono uppercase border ${colorClass}`}>
        {action.replace(/_/g, ' ')}
      </span>
    );
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
      <div className="space-y-6 animate-fadeIn" data-testid="audit-logs-page">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-white uppercase tracking-tight font-['Chivo']">
              Audit Logs
            </h1>
            <p className="text-sm text-zinc-500 font-mono">
              {logs.length} entries found
            </p>
          </div>
          <Button
            variant="ghost"
            onClick={fetchLogs}
            className="text-zinc-400 hover:text-white hover:bg-zinc-800"
            data-testid="refresh-logs-btn"
          >
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
        </div>

        {/* Filters */}
        <Card className="sentinel-card">
          <CardContent className="p-4">
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2 text-zinc-500">
                <Filter className="w-4 h-4" />
                <span className="text-xs font-mono uppercase">Filters</span>
              </div>
              <Select
                value={filters.action}
                onValueChange={(value) => setFilters({...filters, action: value})}
              >
                <SelectTrigger className="w-48 bg-zinc-900 border-zinc-800 text-white" data-testid="action-filter">
                  <SelectValue placeholder="Action" />
                </SelectTrigger>
                <SelectContent className="bg-zinc-900 border-zinc-800">
                  <SelectItem value="all" className="text-white hover:bg-zinc-800">All Actions</SelectItem>
                  <SelectItem value="login_success" className="text-white hover:bg-zinc-800">Login Success</SelectItem>
                  <SelectItem value="login_failed" className="text-white hover:bg-zinc-800">Login Failed</SelectItem>
                  <SelectItem value="user_registered" className="text-white hover:bg-zinc-800">User Registered</SelectItem>
                  <SelectItem value="camera_created" className="text-white hover:bg-zinc-800">Camera Created</SelectItem>
                  <SelectItem value="camera_deleted" className="text-white hover:bg-zinc-800">Camera Deleted</SelectItem>
                  <SelectItem value="settings_updated" className="text-white hover:bg-zinc-800">Settings Updated</SelectItem>
                </SelectContent>
              </Select>
              <Select
                value={filters.resource_type}
                onValueChange={(value) => setFilters({...filters, resource_type: value})}
              >
                <SelectTrigger className="w-40 bg-zinc-900 border-zinc-800 text-white" data-testid="resource-filter">
                  <SelectValue placeholder="Resource" />
                </SelectTrigger>
                <SelectContent className="bg-zinc-900 border-zinc-800">
                  <SelectItem value="all" className="text-white hover:bg-zinc-800">All Resources</SelectItem>
                  <SelectItem value="auth" className="text-white hover:bg-zinc-800">Auth</SelectItem>
                  <SelectItem value="user" className="text-white hover:bg-zinc-800">User</SelectItem>
                  <SelectItem value="camera" className="text-white hover:bg-zinc-800">Camera</SelectItem>
                  <SelectItem value="settings" className="text-white hover:bg-zinc-800">Settings</SelectItem>
                  <SelectItem value="event" className="text-white hover:bg-zinc-800">Event</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>

        {/* Logs Table */}
        <Card className="sentinel-card">
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow className="border-zinc-800 hover:bg-transparent">
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">Time</TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">User</TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">Action</TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">Resource</TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">Details</TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">IP Address</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {logs.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center py-12">
                      <FileText className="w-12 h-12 text-zinc-700 mx-auto mb-3" />
                      <p className="text-zinc-500 font-mono">No audit logs found</p>
                    </TableCell>
                  </TableRow>
                ) : (
                  logs.map((log) => (
                    <TableRow 
                      key={log.id}
                      className="border-zinc-800 hover:bg-zinc-800/50"
                      data-testid={`log-row-${log.id}`}
                    >
                      <TableCell>
                        <div className="flex items-center gap-1.5 text-zinc-400">
                          <Clock className="w-3 h-3" />
                          <span className="text-xs font-mono">
                            {new Date(log.created_at).toLocaleString()}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1.5">
                          <User className="w-3 h-3 text-zinc-500" />
                          <span className="text-sm text-white">{log.username}</span>
                        </div>
                      </TableCell>
                      <TableCell>
                        {getActionBadge(log.action)}
                      </TableCell>
                      <TableCell>
                        <span className="text-xs bg-zinc-800 text-zinc-300 px-2 py-1 rounded-sm font-mono uppercase">
                          {log.resource_type}
                        </span>
                      </TableCell>
                      <TableCell>
                        {log.details ? (
                          <span className="text-xs text-zinc-500 font-mono max-w-xs truncate block">
                            {JSON.stringify(log.details)}
                          </span>
                        ) : (
                          <span className="text-zinc-600">—</span>
                        )}
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1.5 text-zinc-500">
                          <Globe className="w-3 h-3" />
                          <span className="text-xs font-mono">{log.ip_address}</span>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
    </MainLayout>
  );
}
