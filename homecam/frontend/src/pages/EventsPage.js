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
  Bell, 
  AlertTriangle, 
  Info, 
  CheckCircle,
  RefreshCw,
  Filter,
  Check
} from "lucide-react";

export default function EventsPage() {
  const { api, user } = useAuth();
  const [events, setEvents] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [filters, setFilters] = useState({
    severity: "all",
    event_type: "all",
    acknowledged: "all"
  });

  const canAcknowledgeEvents = ["admin", "security_operator"].includes(user?.role);

  useEffect(() => {
    fetchEvents();
  }, [filters]);

  const fetchEvents = async () => {
    try {
      const params = {};
      if (filters.severity !== "all") params.severity = filters.severity;
      if (filters.event_type !== "all") params.event_type = filters.event_type;
      if (filters.acknowledged !== "all") params.acknowledged = filters.acknowledged === "true";
      
      const response = await api.get("/events", { params });
      setEvents(response.data);
    } catch (error) {
      toast.error("Failed to fetch events");
    } finally {
      setIsLoading(false);
    }
  };

  const acknowledgeEvent = async (eventId) => {
    try {
      await api.post(`/events/${eventId}/acknowledge`);
      toast.success("Event acknowledged");
      fetchEvents();
    } catch (error) {
      toast.error("Failed to acknowledge event");
    }
  };

  const getSeverityIcon = (severity) => {
    switch (severity) {
      case "critical":
        return <AlertTriangle className="w-4 h-4 text-red-500" />;
      case "warning":
        return <AlertTriangle className="w-4 h-4 text-yellow-500" />;
      default:
        return <Info className="w-4 h-4 text-blue-500" />;
    }
  };

  const getSeverityBadge = (severity) => {
    switch (severity) {
      case "critical":
        return <span className="badge-critical">Critical</span>;
      case "warning":
        return <span className="badge-warning">Warning</span>;
      default:
        return <span className="text-xs bg-blue-500/10 text-blue-500 border border-blue-500/20 px-2 py-0.5 rounded-sm font-mono uppercase">Info</span>;
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
      <div className="space-y-6 animate-fadeIn" data-testid="events-page">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-white uppercase tracking-tight font-['Chivo']">
              Events & Alarms
            </h1>
            <p className="text-sm text-zinc-500 font-mono">
              {events.length} events found
            </p>
          </div>
          <Button
            variant="ghost"
            onClick={fetchEvents}
            className="text-zinc-400 hover:text-white hover:bg-zinc-800"
            data-testid="refresh-events-btn"
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
                value={filters.severity}
                onValueChange={(value) => setFilters({...filters, severity: value})}
              >
                <SelectTrigger className="w-40 bg-zinc-900 border-zinc-800 text-white" data-testid="severity-filter">
                  <SelectValue placeholder="Severity" />
                </SelectTrigger>
                <SelectContent className="bg-zinc-900 border-zinc-800">
                  <SelectItem value="all" className="text-white hover:bg-zinc-800">All Severities</SelectItem>
                  <SelectItem value="critical" className="text-white hover:bg-zinc-800">Critical</SelectItem>
                  <SelectItem value="warning" className="text-white hover:bg-zinc-800">Warning</SelectItem>
                  <SelectItem value="info" className="text-white hover:bg-zinc-800">Info</SelectItem>
                </SelectContent>
              </Select>
              <Select
                value={filters.event_type}
                onValueChange={(value) => setFilters({...filters, event_type: value})}
              >
                <SelectTrigger className="w-40 bg-zinc-900 border-zinc-800 text-white" data-testid="type-filter">
                  <SelectValue placeholder="Type" />
                </SelectTrigger>
                <SelectContent className="bg-zinc-900 border-zinc-800">
                  <SelectItem value="all" className="text-white hover:bg-zinc-800">All Types</SelectItem>
                  <SelectItem value="motion" className="text-white hover:bg-zinc-800">Motion</SelectItem>
                  <SelectItem value="alarm" className="text-white hover:bg-zinc-800">Alarm</SelectItem>
                  <SelectItem value="connection_lost" className="text-white hover:bg-zinc-800">Connection Lost</SelectItem>
                  <SelectItem value="connection_restored" className="text-white hover:bg-zinc-800">Connection Restored</SelectItem>
                </SelectContent>
              </Select>
              <Select
                value={filters.acknowledged}
                onValueChange={(value) => setFilters({...filters, acknowledged: value})}
              >
                <SelectTrigger className="w-40 bg-zinc-900 border-zinc-800 text-white" data-testid="acknowledged-filter">
                  <SelectValue placeholder="Status" />
                </SelectTrigger>
                <SelectContent className="bg-zinc-900 border-zinc-800">
                  <SelectItem value="all" className="text-white hover:bg-zinc-800">All Status</SelectItem>
                  <SelectItem value="false" className="text-white hover:bg-zinc-800">Pending</SelectItem>
                  <SelectItem value="true" className="text-white hover:bg-zinc-800">Acknowledged</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>

        {/* Events Table */}
        <Card className="sentinel-card">
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow className="border-zinc-800 hover:bg-transparent">
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase w-8"></TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">Severity</TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">Type</TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">Camera</TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">Message</TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">Time</TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">Status</TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {events.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={8} className="text-center py-12">
                      <Bell className="w-12 h-12 text-zinc-700 mx-auto mb-3" />
                      <p className="text-zinc-500 font-mono">No events found</p>
                    </TableCell>
                  </TableRow>
                ) : (
                  events.map((event) => (
                    <TableRow 
                      key={event.id}
                      className={`border-zinc-800 hover:bg-zinc-800/50 ${
                        !event.acknowledged && event.severity === 'critical' 
                          ? 'bg-red-500/5' 
                          : !event.acknowledged && event.severity === 'warning'
                          ? 'bg-yellow-500/5'
                          : ''
                      }`}
                      data-testid={`event-row-${event.id}`}
                    >
                      <TableCell>
                        {getSeverityIcon(event.severity)}
                      </TableCell>
                      <TableCell>
                        {getSeverityBadge(event.severity)}
                      </TableCell>
                      <TableCell>
                        <span className="text-xs bg-zinc-800 text-zinc-300 px-2 py-1 rounded-sm font-mono uppercase">
                          {event.event_type.replace('_', ' ')}
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm text-white">{event.camera_name}</span>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm text-zinc-300 max-w-xs truncate block">
                          {event.message}
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="text-xs text-zinc-500 font-mono">
                          {new Date(event.created_at).toLocaleString()}
                        </span>
                      </TableCell>
                      <TableCell>
                        {event.acknowledged ? (
                          <div className="flex items-center gap-1.5">
                            <CheckCircle className="w-3.5 h-3.5 text-emerald-500" />
                            <span className="text-xs text-emerald-500 font-mono">ACK</span>
                          </div>
                        ) : (
                          <span className="text-xs text-yellow-500 font-mono animate-pulse-slow">PENDING</span>
                        )}
                      </TableCell>
                      <TableCell className="text-right">
                        {!event.acknowledged && canAcknowledgeEvents && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => acknowledgeEvent(event.id)}
                            className="text-zinc-400 hover:text-emerald-500 hover:bg-emerald-500/10"
                            data-testid={`acknowledge-${event.id}`}
                          >
                            <Check className="w-4 h-4" />
                          </Button>
                        )}
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
