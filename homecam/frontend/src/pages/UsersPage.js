import { useState, useEffect } from "react";
import { MainLayout } from "@/components/layout/MainLayout";
import { useAuth } from "@/contexts/AuthContext";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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
  Users, 
  Plus, 
  Shield, 
  ShieldCheck,
  Eye,
  Trash2,
  Clock,
  Mail
} from "lucide-react";

export default function UsersPage() {
  const { api, user: currentUser } = useAuth();
  const [users, setUsers] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showAddDialog, setShowAddDialog] = useState(false);
  const [deleteUser, setDeleteUser] = useState(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  
  const [formData, setFormData] = useState({
    username: "",
    email: "",
    password: "",
    role: "viewer"
  });

  useEffect(() => {
    fetchUsers();
  }, []);

  const fetchUsers = async () => {
    try {
      const response = await api.get("/users");
      setUsers(response.data);
    } catch (error) {
      toast.error("Failed to fetch users");
    } finally {
      setIsLoading(false);
    }
  };

  const handleAddUser = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    
    try {
      await api.post("/auth/register", formData);
      toast.success("User created successfully");
      setShowAddDialog(false);
      resetForm();
      fetchUsers();
    } catch (error) {
      const message = error.response?.data?.detail || "Failed to create user";
      toast.error(message);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDeleteUser = async () => {
    if (!deleteUser) return;
    
    try {
      await api.delete(`/users/${deleteUser.id}`);
      toast.success("User deleted successfully");
      setDeleteUser(null);
      fetchUsers();
    } catch (error) {
      toast.error("Failed to delete user");
    }
  };

  const handleRoleChange = async (userId, newRole) => {
    try {
      await api.put(`/users/${userId}/role`, null, { params: { role: newRole } });
      toast.success("Role updated successfully");
      fetchUsers();
    } catch (error) {
      toast.error("Failed to update role");
    }
  };

  const resetForm = () => {
    setFormData({
      username: "",
      email: "",
      password: "",
      role: "viewer"
    });
  };

  const getRoleIcon = (role) => {
    switch (role) {
      case "admin":
        return <ShieldCheck className="w-4 h-4 text-cyan-500" />;
      case "security_operator":
        return <Shield className="w-4 h-4 text-yellow-500" />;
      default:
        return <Eye className="w-4 h-4 text-zinc-500" />;
    }
  };

  const getRoleBadge = (role) => {
    switch (role) {
      case "admin":
        return (
          <span className="text-xs bg-cyan-500/10 text-cyan-500 border border-cyan-500/20 px-2 py-0.5 rounded-sm font-mono uppercase">
            Admin
          </span>
        );
      case "security_operator":
        return (
          <span className="text-xs bg-yellow-500/10 text-yellow-500 border border-yellow-500/20 px-2 py-0.5 rounded-sm font-mono uppercase">
            Operator
          </span>
        );
      default:
        return (
          <span className="text-xs bg-zinc-500/10 text-zinc-500 border border-zinc-500/20 px-2 py-0.5 rounded-sm font-mono uppercase">
            Viewer
          </span>
        );
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
      <div className="space-y-6 animate-fadeIn" data-testid="users-page">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-white uppercase tracking-tight font-['Chivo']">
              User Management
            </h1>
            <p className="text-sm text-zinc-500 font-mono">
              {users.length} users registered
            </p>
          </div>
          <Dialog open={showAddDialog} onOpenChange={setShowAddDialog}>
            <DialogTrigger asChild>
              <Button className="sentinel-btn-primary" data-testid="add-user-btn">
                <Plus className="w-4 h-4 mr-2" />
                Add User
              </Button>
            </DialogTrigger>
            <DialogContent className="bg-zinc-900 border-zinc-800 max-w-md">
              <DialogHeader>
                <DialogTitle className="text-xl font-bold text-white uppercase tracking-tight">
                  Create New User
                </DialogTitle>
                <DialogDescription className="text-zinc-500 font-mono text-xs">
                  Add a new operator to the system
                </DialogDescription>
              </DialogHeader>
              <form onSubmit={handleAddUser} className="space-y-4">
                <div className="space-y-2">
                  <Label className="text-zinc-400 text-xs uppercase tracking-wider">
                    Username
                  </Label>
                  <Input
                    value={formData.username}
                    onChange={(e) => setFormData({...formData, username: e.target.value})}
                    placeholder="Enter username"
                    className="sentinel-input"
                    required
                    minLength={3}
                    data-testid="new-user-username"
                  />
                </div>
                <div className="space-y-2">
                  <Label className="text-zinc-400 text-xs uppercase tracking-wider">
                    Email
                  </Label>
                  <Input
                    type="email"
                    value={formData.email}
                    onChange={(e) => setFormData({...formData, email: e.target.value})}
                    placeholder="user@example.com"
                    className="sentinel-input"
                    required
                    data-testid="new-user-email"
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
                    placeholder="Min 12 characters"
                    className="sentinel-input"
                    required
                    minLength={12}
                    data-testid="new-user-password"
                  />
                </div>
                <div className="space-y-2">
                  <Label className="text-zinc-400 text-xs uppercase tracking-wider">
                    Role
                  </Label>
                  <Select
                    value={formData.role}
                    onValueChange={(value) => setFormData({...formData, role: value})}
                  >
                    <SelectTrigger className="bg-zinc-950 border-zinc-800 text-white" data-testid="new-user-role">
                      <SelectValue placeholder="Select role" />
                    </SelectTrigger>
                    <SelectContent className="bg-zinc-900 border-zinc-800">
                      <SelectItem value="viewer" className="text-white hover:bg-zinc-800">Viewer</SelectItem>
                      <SelectItem value="security_operator" className="text-white hover:bg-zinc-800">Security Operator</SelectItem>
                      <SelectItem value="admin" className="text-white hover:bg-zinc-800">Admin</SelectItem>
                    </SelectContent>
                  </Select>
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
                    data-testid="submit-user-btn"
                  >
                    {isSubmitting ? "Creating..." : "Create User"}
                  </Button>
                </DialogFooter>
              </form>
            </DialogContent>
          </Dialog>
        </div>

        {/* Users Table */}
        <Card className="sentinel-card">
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow className="border-zinc-800 hover:bg-transparent">
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">User</TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">Email</TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">Role</TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">2FA</TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase">Last Login</TableHead>
                  <TableHead className="text-zinc-400 font-mono text-xs uppercase text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {users.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={6} className="text-center py-12">
                      <Users className="w-12 h-12 text-zinc-700 mx-auto mb-3" />
                      <p className="text-zinc-500 font-mono">No users found</p>
                    </TableCell>
                  </TableRow>
                ) : (
                  users.map((user) => (
                    <TableRow 
                      key={user.id}
                      className="border-zinc-800 hover:bg-zinc-800/50"
                      data-testid={`user-row-${user.id}`}
                    >
                      <TableCell>
                        <div className="flex items-center gap-3">
                          <div className="w-8 h-8 bg-zinc-800 rounded-sm flex items-center justify-center">
                            {getRoleIcon(user.role)}
                          </div>
                          <div>
                            <p className="text-white font-medium">{user.username}</p>
                            {user.id === currentUser?.id && (
                              <span className="text-xs text-cyan-500 font-mono">(You)</span>
                            )}
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1.5 text-zinc-400">
                          <Mail className="w-3 h-3" />
                          <span className="text-sm">{user.email}</span>
                        </div>
                      </TableCell>
                      <TableCell>
                        {user.id === currentUser?.id ? (
                          getRoleBadge(user.role)
                        ) : (
                          <Select
                            value={user.role}
                            onValueChange={(value) => handleRoleChange(user.id, value)}
                          >
                            <SelectTrigger className="w-36 bg-zinc-900 border-zinc-800 text-white h-8">
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent className="bg-zinc-900 border-zinc-800">
                              <SelectItem value="viewer" className="text-white hover:bg-zinc-800">Viewer</SelectItem>
                              <SelectItem value="security_operator" className="text-white hover:bg-zinc-800">Operator</SelectItem>
                              <SelectItem value="admin" className="text-white hover:bg-zinc-800">Admin</SelectItem>
                            </SelectContent>
                          </Select>
                        )}
                      </TableCell>
                      <TableCell>
                        {user.totp_enabled ? (
                          <span className="badge-online">Enabled</span>
                        ) : (
                          <span className="badge-warning">Disabled</span>
                        )}
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1.5 text-zinc-500">
                          <Clock className="w-3 h-3" />
                          <span className="text-xs font-mono">
                            {user.last_login 
                              ? new Date(user.last_login).toLocaleString()
                              : "Never"
                            }
                          </span>
                        </div>
                      </TableCell>
                      <TableCell className="text-right">
                        {user.id !== currentUser?.id && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setDeleteUser(user)}
                            className="text-zinc-400 hover:text-red-500 hover:bg-red-500/10"
                            data-testid={`delete-user-${user.id}`}
                          >
                            <Trash2 className="w-4 h-4" />
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

        {/* Delete Confirmation Dialog */}
        <AlertDialog open={!!deleteUser} onOpenChange={() => setDeleteUser(null)}>
          <AlertDialogContent className="bg-zinc-900 border-zinc-800">
            <AlertDialogHeader>
              <AlertDialogTitle className="text-white">Delete User</AlertDialogTitle>
              <AlertDialogDescription className="text-zinc-400">
                Are you sure you want to delete "{deleteUser?.username}"? This action cannot be undone.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel className="bg-zinc-800 border-zinc-700 text-white hover:bg-zinc-700">
                Cancel
              </AlertDialogCancel>
              <AlertDialogAction 
                onClick={handleDeleteUser}
                className="sentinel-btn-destructive"
                data-testid="confirm-delete-user"
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
