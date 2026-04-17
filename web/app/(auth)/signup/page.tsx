'use client';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { toast } from 'sonner';
import { Box, Button, Flex, Heading, Text, TextField } from '@radix-ui/themes';
import { ArrowRightIcon } from '@radix-ui/react-icons';
import { Spinner } from '@/components/common/spinner';
import { signupSchema, type SignupInput } from '@/lib/schemas';
import { api, ApiError } from '@/lib/api';

export default function SignupPage() {
  const router = useRouter();
  const form = useForm<SignupInput>({
    resolver: zodResolver(signupSchema),
    defaultValues: { email: '', password: '' },
  });

  async function onSubmit(values: SignupInput) {
    try {
      await api('/api/auth/signup', { method: 'POST', body: JSON.stringify(values) });
      router.push('/');
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : 'signup failed');
    }
  }

  return (
    <Box>
      <Heading size="6" mb="1">
        Create your TestPay account
      </Heading>
      <Text size="2" color="gray" mb="5" as="p">
        Start mocking payment gateways in under a minute.
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
              autoComplete="new-password"
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
            className="transition-transform hover:-translate-y-px"
          >
            {form.formState.isSubmitting && <Spinner size="small" />}
            {form.formState.isSubmitting ? 'Creating…' : 'Create account'}
          </Button>
          <Text size="2" color="gray" align="center">
            Already have an account?{' '}
            <Link
              href="/login"
              className="group inline-flex items-center gap-1 font-medium text-[var(--accent-11)] transition-colors hover:text-[var(--accent-12)]"
            >
              Sign in
              <ArrowRightIcon className="h-3 w-3 transition-transform duration-200 group-hover:translate-x-0.5" />
            </Link>
          </Text>
        </Flex>
      </form>
    </Box>
  );
}
