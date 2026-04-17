'use client';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { toast } from 'sonner';
import { Box, Button, Flex, Heading, Text, TextField } from '@radix-ui/themes';
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
    <Box>
      <Heading size="6" mb="1">
        Sign in to TestPay
      </Heading>
      <Text size="2" color="gray" mb="5" as="p">
        Welcome back. Enter your credentials to continue.
      </Text>
      <form onSubmit={form.handleSubmit(onSubmit)}>
        <Flex direction="column" gap="4">
          <Box>
            <Text as="label" size="2" weight="medium" htmlFor="email">
              Email
            </Text>
            <TextField.Root
              id="email"
              type="email"
              autoComplete="email"
              mt="1"
              {...form.register('email')}
            />
            {form.formState.errors.email && (
              <Text size="1" color="red" mt="1" as="div">
                {form.formState.errors.email.message}
              </Text>
            )}
          </Box>
          <Box>
            <Text as="label" size="2" weight="medium" htmlFor="password">
              Password
            </Text>
            <TextField.Root
              id="password"
              type="password"
              autoComplete="current-password"
              mt="1"
              {...form.register('password')}
            />
            {form.formState.errors.password && (
              <Text size="1" color="red" mt="1" as="div">
                {form.formState.errors.password.message}
              </Text>
            )}
          </Box>
          <Button
            type="submit"
            size="3"
            disabled={form.formState.isSubmitting}
            loading={form.formState.isSubmitting}
          >
            {form.formState.isSubmitting ? 'Signing in…' : 'Sign in'}
          </Button>
          <Text size="2" color="gray" align="center">
            No account?{' '}
            <Link href="/signup" className="underline text-[var(--accent-11)]">
              Create one
            </Link>
          </Text>
        </Flex>
      </form>
    </Box>
  );
}
