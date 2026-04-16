'use client';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { toast } from 'sonner';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { loginSchema, type LoginInput } from '@/lib/schemas';
import { api, ApiError } from '@/lib/api';

export default function LoginPage() {
  const router = useRouter();
  const form = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: '', password: '' },
  });

  async function onSubmit(values: LoginInput) {
    try {
      await api('/api/auth/login', { method: 'POST', body: JSON.stringify(values) });
      router.push('/');
    } catch (e) {
      const msg = e instanceof ApiError ? e.message : 'login failed';
      toast.error(msg);
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Sign in to TestPay</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
          <div>
            <Label htmlFor="email">Email</Label>
            <Input id="email" type="email" autoComplete="email" {...form.register('email')} />
            {form.formState.errors.email && (
              <p className="text-xs text-destructive mt-1">{form.formState.errors.email.message}</p>
            )}
          </div>
          <div>
            <Label htmlFor="password">Password</Label>
            <Input
              id="password"
              type="password"
              autoComplete="current-password"
              {...form.register('password')}
            />
            {form.formState.errors.password && (
              <p className="text-xs text-destructive mt-1">
                {form.formState.errors.password.message}
              </p>
            )}
          </div>
          <Button type="submit" className="w-full" disabled={form.formState.isSubmitting}>
            {form.formState.isSubmitting ? 'Signing in…' : 'Sign in'}
          </Button>
          <p className="text-xs text-center text-muted-foreground">
            No account?{' '}
            <Link href="/signup" className="text-foreground underline">
              Create one
            </Link>
          </p>
        </form>
      </CardContent>
    </Card>
  );
}
