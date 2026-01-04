# UI Standards - Dark Tech Theme

> **React + TypeScript + Vite + Tailwind CSS + Zustand**

This document defines the UI standards for generated React applications. All generated UIs follow the **Dark Tech Theme** with cyan accents.

---

## Tech Stack

| Technology | Version | Purpose |
|------------|---------|---------|
| **React** | 18.x | UI framework |
| **TypeScript** | 5.x | Type safety |
| **Vite** | 5.x | Build tool |
| **Tailwind CSS** | 3.4+ | Styling |
| **Zustand** | 5.x | State management |
| **Recharts** | 3.x | Charts and visualizations |
| **Lucide React** | Latest | Icons |
| **Axios** | 1.x | HTTP client |
| **React Flow** | 11.x | Node-based diagrams (optional) |

---

## Color Palette

### Primary Colors

```css
/* Dark Tech Theme - Cyan Accents */

/* Backgrounds */
--bg-primary: #0f172a;      /* slate-900 */
--bg-secondary: #1e293b;    /* slate-800 */
--bg-tertiary: #334155;     /* slate-700 */
--bg-glass: rgba(30, 41, 59, 0.3);  /* slate-800/30 */

/* Primary Accent - Cyan */
--primary-400: #22d3ee;     /* cyan-400 */
--primary-500: #06b6d4;     /* cyan-500 */
--primary-600: #0891b2;     /* cyan-600 */

/* Secondary Accent - Blue */
--secondary-500: #3b82f6;   /* blue-500 */
--secondary-600: #2563eb;   /* blue-600 */

/* Status Colors */
--success: #34d399;         /* emerald-400 */
--warning: #fbbf24;         /* yellow-400 */
--error: #f87171;           /* red-400 */

/* Text */
--text-primary: #ffffff;
--text-secondary: #94a3b8;  /* slate-400 */
--text-muted: #64748b;      /* slate-500 */
```

### Tailwind Usage

```jsx
// Background with glassmorphism
<div className="backdrop-blur-xl bg-slate-800/30 border border-cyan-500/20">

// Primary button
<button className="bg-gradient-to-r from-cyan-500 to-blue-600 hover:from-cyan-400 hover:to-blue-500">

// Card component
<div className="rounded-lg bg-slate-800/50 border border-cyan-500/20 shadow-lg">

// Status indicators
<span className="text-emerald-400">Online</span>
<span className="text-yellow-400">Warning</span>
<span className="text-red-400">Error</span>
```

---

## Base CSS (index.css)

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

/* Disable ligatures for monospace fonts */
.font-mono, code, pre, [class*="font-mono"] {
  font-variant-ligatures: none;
  font-feature-settings: "liga" 0, "clig" 0;
}

/* Global scrollbar styles - Cyan theme */
* {
  scrollbar-width: thin;
  scrollbar-color: rgba(6, 182, 212, 0.6) rgba(15, 23, 42, 0.4);
}

*::-webkit-scrollbar {
  width: 10px;
  height: 10px;
}

*::-webkit-scrollbar-track {
  background: rgba(15, 23, 42, 0.4);
  border-radius: 10px;
  border: 1px solid rgba(6, 182, 212, 0.1);
}

*::-webkit-scrollbar-thumb {
  background: linear-gradient(180deg, rgba(6, 182, 212, 0.8), rgba(8, 145, 178, 0.6));
  border-radius: 10px;
  border: 1px solid rgba(6, 182, 212, 0.3);
}

*::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(180deg, #06b6d4, #0891b2);
  box-shadow: 0 0 10px rgba(6, 182, 212, 0.5);
}

/* Pulse animation for loading/working states */
@keyframes pulse-glow {
  0%, 100% { box-shadow: 0 0 10px currentColor; }
  50% { box-shadow: 0 0 20px currentColor, 0 0 30px currentColor; }
}

.pulse-glow {
  animation: pulse-glow 2s ease-in-out infinite;
}

/* Prevent overscroll bounce */
body {
  overscroll-behavior: none;
  background-color: #0f172a;
  color: #ffffff;
}
```

---

## Component Patterns

### Layout Structure

```
┌─────────────────────────────────────────────────────────────┐
│  Header (h-16/h-20, fixed top, glassmorphism)               │
├────────────┬────────────────────────────────────────────────┤
│            │                                                │
│  Sidebar   │              Main Content                      │
│  (w-64)    │              (flex-1)                          │
│  Collaps-  │                                                │
│  ible      │                                                │
│            │                                                │
└────────────┴────────────────────────────────────────────────┘
```

```tsx
// App.tsx structure
function App() {
  const [activeView, setActiveView] = useState("dashboard");

  return (
    <div className="min-h-screen bg-slate-950">
      <Header />
      <div className="flex pt-16">
        <Sidebar activeView={activeView} onViewChange={setActiveView} />
        <main className="flex-1 p-6 ml-64">
          {renderView(activeView)}
        </main>
      </div>
    </div>
  );
}
```

### Header Component

```tsx
interface StatCardProps {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  color?: string;
}

const StatCard = ({ icon, label, value, color = "text-white" }: StatCardProps) => (
  <div className="flex items-center gap-3 px-3 py-2 rounded-lg backdrop-blur-sm bg-slate-800/20 border border-cyan-500/20 hover:bg-slate-800/30 transition-all">
    <div className="text-cyan-400">{icon}</div>
    <div>
      <p className="text-xs text-gray-400">{label}</p>
      <p className={`text-lg font-bold ${color} font-mono`}>{value}</p>
    </div>
  </div>
);

export default function Header() {
  return (
    <header className="fixed top-0 left-0 right-0 h-16 backdrop-blur-xl bg-slate-900/30 border-b border-cyan-500/30 z-40 shadow-lg">
      <div className="flex items-center justify-between h-full px-6">
        {/* Logo */}
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-cyan-500 to-blue-600 flex items-center justify-center">
            <Server className="w-6 h-6 text-white" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-white">App Name</h1>
            <p className="text-xs text-cyan-400 font-mono">Subtitle</p>
          </div>
        </div>
        
        {/* Stats */}
        <div className="flex items-center gap-4">
          <StatCard icon={<Activity />} label="Active" value={42} />
          <StatCard icon={<AlertTriangle />} label="Alerts" value={3} color="text-red-400" />
        </div>
      </div>
    </header>
  );
}
```

### Sidebar Component

```tsx
interface NavItem {
  id: string;
  label: string;
  icon: React.ReactNode;
}

export default function Sidebar({ activeView, onViewChange }) {
  const navItems: NavItem[] = [
    { id: "dashboard", label: "Dashboard", icon: <LayoutDashboard /> },
    { id: "list", label: "Items", icon: <List /> },
    { id: "settings", label: "Settings", icon: <Settings /> },
  ];

  return (
    <aside className="fixed left-0 top-16 bottom-0 w-64 backdrop-blur-xl bg-slate-900/50 border-r border-cyan-500/20 z-30">
      <nav className="p-4 space-y-2">
        {navItems.map((item) => (
          <button
            key={item.id}
            onClick={() => onViewChange(item.id)}
            className={`w-full flex items-center gap-3 px-4 py-3 rounded-lg transition-all ${
              activeView === item.id
                ? "bg-cyan-500/20 text-cyan-400 border border-cyan-500/30"
                : "text-gray-400 hover:bg-slate-800/50 hover:text-white"
            }`}
          >
            {item.icon}
            <span className="font-medium">{item.label}</span>
          </button>
        ))}
      </nav>
    </aside>
  );
}
```

### Card Component

```tsx
interface CardProps {
  title: string;
  children: React.ReactNode;
  className?: string;
}

export function Card({ title, children, className = "" }: CardProps) {
  return (
    <div className={`rounded-lg backdrop-blur-sm bg-slate-800/30 border border-cyan-500/20 ${className}`}>
      <div className="px-4 py-3 border-b border-cyan-500/10">
        <h3 className="text-lg font-semibold text-white">{title}</h3>
      </div>
      <div className="p-4">
        {children}
      </div>
    </div>
  );
}
```

### Table Component

```tsx
interface Column<T> {
  key: keyof T;
  label: string;
  render?: (value: T[keyof T], row: T) => React.ReactNode;
}

interface DataTableProps<T> {
  columns: Column<T>[];
  data: T[];
  onRowClick?: (row: T) => void;
}

export function DataTable<T>({ columns, data, onRowClick }: DataTableProps<T>) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full">
        <thead>
          <tr className="border-b border-cyan-500/20">
            {columns.map((col) => (
              <th key={String(col.key)} className="px-4 py-3 text-left text-sm font-semibold text-gray-400">
                {col.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.map((row, i) => (
            <tr
              key={i}
              onClick={() => onRowClick?.(row)}
              className="border-b border-slate-700/50 hover:bg-slate-800/30 cursor-pointer transition-colors"
            >
              {columns.map((col) => (
                <td key={String(col.key)} className="px-4 py-3 text-sm text-white">
                  {col.render ? col.render(row[col.key], row) : String(row[col.key])}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
```

### Button Variants

```tsx
type ButtonVariant = "primary" | "secondary" | "danger" | "ghost";

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  loading?: boolean;
}

const variantStyles: Record<ButtonVariant, string> = {
  primary: "bg-gradient-to-r from-cyan-500 to-blue-600 hover:from-cyan-400 hover:to-blue-500 text-white",
  secondary: "bg-slate-700 hover:bg-slate-600 text-white border border-slate-600",
  danger: "bg-red-500/20 hover:bg-red-500/30 text-red-400 border border-red-500/30",
  ghost: "hover:bg-slate-800/50 text-gray-400 hover:text-white",
};

export function Button({ variant = "primary", loading, children, className = "", ...props }: ButtonProps) {
  return (
    <button
      className={`px-4 py-2 rounded-lg font-medium transition-all disabled:opacity-50 ${variantStyles[variant]} ${className}`}
      disabled={loading}
      {...props}
    >
      {loading ? <Loader2 className="w-4 h-4 animate-spin" /> : children}
    </button>
  );
}
```

### Status Badge

```tsx
type Status = "success" | "warning" | "error" | "info" | "neutral";

const statusStyles: Record<Status, string> = {
  success: "bg-emerald-500/20 text-emerald-400 border-emerald-500/30",
  warning: "bg-yellow-500/20 text-yellow-400 border-yellow-500/30",
  error: "bg-red-500/20 text-red-400 border-red-500/30",
  info: "bg-cyan-500/20 text-cyan-400 border-cyan-500/30",
  neutral: "bg-slate-500/20 text-slate-400 border-slate-500/30",
};

export function StatusBadge({ status, children }: { status: Status; children: React.ReactNode }) {
  return (
    <span className={`px-2 py-1 rounded-full text-xs font-medium border ${statusStyles[status]}`}>
      {children}
    </span>
  );
}
```

---

## State Management (Zustand)

### Store Pattern

```tsx
// stores/entityStore.ts
import { create } from "zustand";

interface Entity {
  id: string;
  name: string;
  status: string;
}

interface EntityStore {
  // State
  entities: Entity[];
  selectedId: string | null;
  isLoading: boolean;
  error: string | null;
  
  // Actions
  setEntities: (entities: Entity[]) => void;
  addEntity: (entity: Entity) => void;
  updateEntity: (id: string, updates: Partial<Entity>) => void;
  deleteEntity: (id: string) => void;
  selectEntity: (id: string | null) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
}

export const useEntityStore = create<EntityStore>((set) => ({
  // Initial state
  entities: [],
  selectedId: null,
  isLoading: false,
  error: null,
  
  // Actions
  setEntities: (entities) => set({ entities }),
  
  addEntity: (entity) => set((state) => ({
    entities: [...state.entities, entity],
  })),
  
  updateEntity: (id, updates) => set((state) => ({
    entities: state.entities.map((e) =>
      e.id === id ? { ...e, ...updates } : e
    ),
  })),
  
  deleteEntity: (id) => set((state) => ({
    entities: state.entities.filter((e) => e.id !== id),
  })),
  
  selectEntity: (id) => set({ selectedId: id }),
  
  setLoading: (isLoading) => set({ isLoading }),
  
  setError: (error) => set({ error }),
}));
```

### Using Store in Components

```tsx
import { useEntityStore } from "../stores/entityStore";

function EntityList() {
  const { entities, isLoading, selectEntity } = useEntityStore();

  if (isLoading) return <LoadingSpinner />;

  return (
    <ul>
      {entities.map((entity) => (
        <li key={entity.id} onClick={() => selectEntity(entity.id)}>
          {entity.name}
        </li>
      ))}
    </ul>
  );
}
```

---

## API Services

### API Client Setup

```tsx
// services/api.ts
import axios from "axios";

const getBaseUrl = () => {
  if (import.meta.env.DEV) {
    return "http://localhost:8080/api";
  }
  return window.location.origin + "/api";
};

export const api = axios.create({
  baseURL: getBaseUrl(),
  timeout: 30000,
  headers: {
    "Content-Type": "application/json",
  },
});

// Request interceptor for auth
api.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor for errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem("token");
      window.location.href = "/login";
    }
    return Promise.reject(error);
  }
);
```

### Service Functions

```tsx
// services/entityService.ts
import { api } from "./api";

export interface Entity {
  id: string;
  name: string;
  status: string;
}

export const entityService = {
  getAll: () => api.get<Entity[]>("/entities"),
  getById: (id: string) => api.get<Entity>(`/entities/${id}`),
  create: (data: Omit<Entity, "id">) => api.post<Entity>("/entities", data),
  update: (id: string, data: Partial<Entity>) => api.put<Entity>(`/entities/${id}`, data),
  delete: (id: string) => api.delete(`/entities/${id}`),
};
```

---

## WebSocket Integration

```tsx
// services/websocket.ts
class WebSocketService {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private listeners: Map<string, ((data: any) => void)[]> = new Map();

  connect() {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsUrl = `${protocol}//${window.location.host}/ws`;

    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      console.log("WebSocket connected");
      this.reconnectAttempts = 0;
    };

    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      const listeners = this.listeners.get(message.type) || [];
      listeners.forEach((callback) => callback(message.data));
    };

    this.ws.onclose = () => {
      if (this.reconnectAttempts < this.maxReconnectAttempts) {
        setTimeout(() => {
          this.reconnectAttempts++;
          this.connect();
        }, 2000 * this.reconnectAttempts);
      }
    };
  }

  subscribe(type: string, callback: (data: any) => void) {
    const listeners = this.listeners.get(type) || [];
    this.listeners.set(type, [...listeners, callback]);
    
    return () => {
      const updated = this.listeners.get(type)?.filter((cb) => cb !== callback) || [];
      this.listeners.set(type, updated);
    };
  }

  disconnect() {
    this.ws?.close();
  }
}

export const websocketService = new WebSocketService();
```

---

## Charts (Recharts)

### Line Chart

```tsx
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from "recharts";

interface DataPoint {
  timestamp: string;
  value: number;
}

export function MetricChart({ data, label }: { data: DataPoint[]; label: string }) {
  return (
    <ResponsiveContainer width="100%" height={300}>
      <LineChart data={data}>
        <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
        <XAxis 
          dataKey="timestamp" 
          stroke="#64748b"
          tick={{ fill: "#94a3b8" }}
        />
        <YAxis 
          stroke="#64748b"
          tick={{ fill: "#94a3b8" }}
        />
        <Tooltip
          contentStyle={{
            backgroundColor: "#1e293b",
            border: "1px solid rgba(6, 182, 212, 0.3)",
            borderRadius: "8px",
          }}
          labelStyle={{ color: "#94a3b8" }}
        />
        <Line
          type="monotone"
          dataKey="value"
          name={label}
          stroke="#06b6d4"
          strokeWidth={2}
          dot={false}
          activeDot={{ r: 6, fill: "#06b6d4" }}
        />
      </LineChart>
    </ResponsiveContainer>
  );
}
```

---

## File Structure

```
frontend/
├── public/
│   └── config.js           # Runtime configuration
├── src/
│   ├── components/
│   │   ├── Header.tsx
│   │   ├── Sidebar.tsx
│   │   ├── Card.tsx
│   │   ├── DataTable.tsx
│   │   ├── Button.tsx
│   │   ├── StatusBadge.tsx
│   │   └── [Feature]Dashboard.tsx
│   ├── services/
│   │   ├── api.ts
│   │   ├── websocket.ts
│   │   └── [entity]Service.ts
│   ├── stores/
│   │   └── [entity]Store.ts
│   ├── types/
│   │   └── index.ts
│   ├── utils/
│   │   └── formatters.ts
│   ├── App.tsx
│   ├── main.tsx
│   └── index.css
├── index.html
├── package.json
├── tailwind.config.js
├── tsconfig.json
└── vite.config.ts
```

---

## Package.json Template

```json
{
  "name": "app-frontend",
  "private": true,
  "version": "1.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "axios": "^1.6.0",
    "lucide-react": "^0.400.0",
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "recharts": "^2.12.0",
    "zustand": "^4.5.0"
  },
  "devDependencies": {
    "@types/react": "^18.2.0",
    "@types/react-dom": "^18.2.0",
    "@vitejs/plugin-react": "^4.2.0",
    "autoprefixer": "^10.4.0",
    "postcss": "^8.4.0",
    "tailwindcss": "^3.4.0",
    "typescript": "^5.3.0",
    "vite": "^5.0.0"
  }
}
```

---

## Tailwind Config

```js
/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Custom colors if needed
      },
    },
  },
  plugins: [],
}
```
