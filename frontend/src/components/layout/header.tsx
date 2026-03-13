'use client';

import { useAuth } from '@/contexts/auth-context';
import { Menu, Store } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetTrigger, SheetTitle, SheetDescription } from '@/components/ui/sheet';
import { Sidebar } from '@/components/layout/sidebar';
import { useState } from 'react';

interface HeaderProps {
  title?: string;
}

export function Header({ title }: HeaderProps) {
  const { user } = useAuth();
  const [open, setOpen] = useState(false);

  return (
    <header className="flex h-16 shrink-0 items-center justify-between border-b bg-card px-6">
      <div className="flex items-center gap-4 min-w-0 overflow-hidden">
        {/* Mobile Menu */}
        <div className="lg:hidden flex-none">
          <Sheet open={open} onOpenChange={setOpen}>
            <SheetTrigger asChild>
              <Button variant="ghost" size="icon">
                <Menu className="h-5 w-5" />
              </Button>
            </SheetTrigger>
            <SheetContent side="left" className="p-0 w-72 bg-card gap-0">
              <SheetTitle className="sr-only">Navigation Menu</SheetTitle>
              <SheetDescription className="sr-only">
                Mobile navigation menu
              </SheetDescription>
              <div className="h-full flex flex-col">
                <div className="flex h-16 items-center border-b px-6 flex-none">
                  <Store className="h-6 w-6 text-primary mr-2" />
                  <span className="font-bold text-lg">DashPoint</span>
                </div>
                <div className="flex-1 overflow-hidden">
                  <Sidebar onNavigate={() => setOpen(false)} />
                </div>
              </div>
            </SheetContent>
          </Sheet>
        </div>

        {title && <h1 className="text-2xl font-bold tracking-tight truncate min-w-0">{title}</h1>}
      </div>

      <div className="flex items-center gap-4 flex-none ml-2">

        {/* User info */}
        <div className="flex items-center gap-3 min-w-0">
          <div className="h-8 w-8 rounded-full bg-primary flex-none flex items-center justify-center">
            <span className="text-sm font-medium text-primary-foreground">
              {user?.name?.charAt(0).toUpperCase() || 'U'}
            </span>
          </div>
          <div className="hidden sm:block min-w-0 overflow-hidden">
            <p className="text-sm font-medium truncate">{user?.name}</p>
            <p className="text-xs text-muted-foreground capitalize truncate">{user?.role_name}</p>
          </div>
        </div>
      </div>
    </header>
  );
}
