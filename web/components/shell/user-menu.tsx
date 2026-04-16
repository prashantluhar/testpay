'use client';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { LogOut, Settings as SettingsIcon, User as UserIcon, ChevronDown } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { api } from '@/lib/api';
import type { User } from '@/lib/types';

// Initials from an email local part, max 2 chars.
function initials(email: string) {
  const local = email.split('@')[0] || email;
  const parts = local.split(/[._-]+/).filter(Boolean);
  if (parts.length >= 2) return (parts[0][0] + parts[1][0]).toUpperCase();
  return local.slice(0, 2).toUpperCase();
}

export function UserMenu({ user }: { user: User }) {
  const router = useRouter();

  async function signOut() {
    try {
      await api('/api/auth/logout', { method: 'POST' });
    } catch {
      /* ignore */
    }
    router.push('/login');
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button
          type="button"
          className="flex items-center gap-2 rounded-md px-2 py-1 hover:bg-accent/50 transition-colors outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          <span className="flex h-7 w-7 items-center justify-center rounded-full bg-emerald-500/15 text-emerald-500 text-[11px] font-semibold">
            {initials(user.email)}
          </span>
          <span className="text-sm max-w-[160px] truncate hidden md:inline">{user.email}</span>
          <ChevronDown className="h-3.5 w-3.5 text-muted-foreground" />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-56">
        <DropdownMenuLabel className="flex items-center gap-2">
          <UserIcon className="h-4 w-4 text-muted-foreground" />
          <div className="flex flex-col min-w-0">
            <span className="text-xs text-muted-foreground">Signed in as</span>
            <span className="text-sm truncate">{user.email}</span>
          </div>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem asChild>
          <Link href="/settings" className="cursor-pointer">
            <SettingsIcon className="h-4 w-4 mr-2" />
            Settings
          </Link>
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem onSelect={signOut} className="cursor-pointer text-red-500 focus:text-red-500">
          <LogOut className="h-4 w-4 mr-2" />
          Sign out
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
