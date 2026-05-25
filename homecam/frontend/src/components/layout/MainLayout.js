import { Sidebar } from "./Sidebar";

export const MainLayout = ({ children }) => {
  return (
    <div className="flex min-h-screen bg-[#09090b]">
      <Sidebar />
      <main className="flex-1 p-6 overflow-auto">
        {children}
      </main>
    </div>
  );
};
