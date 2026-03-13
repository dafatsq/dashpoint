import {
  LayoutDashboard,
  ShoppingCart,
  Package,
  Boxes,
  Receipt,
  BarChart3,
  Users,
  ScrollText,
  Clock,
  History,
  Settings,
  Wallet,
  Layers,
} from 'lucide-react';
import { PERMISSIONS } from '@/contexts/auth-context';

export interface NavItem {
  href: string;
  label: string;
  icon: React.ReactNode;
  permission?: string;
  permissions?: string[];
  description?: string;
  color?: string;
}

export const navItems: NavItem[] = [
  {
    href: '/',
    label: 'Dashboard',
    icon: <LayoutDashboard className="h-5 w-5" />,
    description: 'Overview of your store performance',
    color: 'text-blue-500',
  },
  {
    href: '/pos',
    label: 'Point of Sale',
    icon: <ShoppingCart className="h-5 w-5" />,
    permission: PERMISSIONS.POS_VIEW,
    description: 'Process sales and manage cart',
    color: 'text-green-500',
  },
  {
    href: '/shifts',
    label: 'Shifts',
    icon: <Clock className="h-5 w-5" />,
    permissions: [PERMISSIONS.POS_VIEW, PERMISSIONS.SALES_VIEW, PERMISSIONS.REPORTS_VIEW],
    description: 'View register shifts and cash drawer',
    color: 'text-amber-500',
  },
  {
    href: '/changes',
    label: 'Recent Changes',
    icon: <History className="h-5 w-5" />,
    description: 'Track price and inventory updates',
    color: 'text-rose-500',
  },
  {
    href: '/products',
    label: 'Products',
    icon: <Package className="h-5 w-5" />,
    permission: PERMISSIONS.PRODUCTS_VIEW,
    description: 'Manage your product catalog',
    color: 'text-orange-500',
  },
  {
    href: '/inventory',
    label: 'Inventory',
    icon: <Boxes className="h-5 w-5" />,
    permission: PERMISSIONS.INVENTORY_VIEW,
    description: 'Track stock and adjustments',
    color: 'text-purple-500',
  },
  {
    href: '/sales',
    label: 'Sales History',
    icon: <Receipt className="h-5 w-5" />,
    permission: PERMISSIONS.SALES_VIEW,
    description: 'View past transactions',
    color: 'text-indigo-500',
  },
  {
    href: '/reports',
    label: 'Reports',
    icon: <BarChart3 className="h-5 w-5" />,
    permission: PERMISSIONS.REPORTS_VIEW,
    description: 'Analyze sales data',
    color: 'text-pink-500',
  },
  {
    href: '/categories',
    label: 'Categories',
    icon: <Layers className="h-4 w-4" />,
    permission: PERMISSIONS.CATEGORIES_VIEW,
    description: 'Manage product and expense categories',
    color: 'text-violet-500',
  },
  {
    href: '/expenses',
    label: 'Expenses',
    icon: <Wallet className="h-5 w-5" />,
    permission: PERMISSIONS.EXPENSES_VIEW,
    description: 'Track business expenses',
    color: 'text-red-500',
  },
  {
    href: '/users',
    label: 'Users',
    icon: <Users className="h-5 w-5" />,
    permission: PERMISSIONS.USERS_VIEW,
    description: 'Manage staff and permissions',
    color: 'text-cyan-500',
  },
  {
    href: '/audit',
    label: 'Audit Logs',
    icon: <ScrollText className="h-5 w-5" />,
    permission: PERMISSIONS.AUDIT_VIEW,
    description: 'View system activity',
    color: 'text-gray-500',
  },

  {
    href: '/settings',
    label: 'Settings',
    icon: <Settings className="h-5 w-5" />,
    description: 'Configure system options',
    color: 'text-slate-500',
  },
];
