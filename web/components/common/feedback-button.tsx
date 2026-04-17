'use client';
import { useState } from 'react';
import {
  Dialog,
  Flex,
  Text,
  Button,
  TextField,
  TextArea,
} from '@radix-ui/themes';
import { ChatBubbleIcon } from '@radix-ui/react-icons';
import { toast } from 'sonner';
import { api, ApiError } from '@/lib/api';

// Small "💬 Feedback" trigger rendered in the dashboard topbar and the docs
// header. Opens a 4-field dialog (what you tried, what worked, what's
// missing, email for follow-up) and POSTs to /api/feedback.
//
// Public endpoint on the backend — works for authenticated and anonymous
// visitors. Logged-in users' workspace/user ids are attached server-side
// from the session cookie.
export function FeedbackButton({ label = 'Feedback', compact = false }: { label?: string; compact?: boolean }) {
  const [open, setOpen] = useState(false);
  const [whatTried, setWhatTried] = useState('');
  const [worked, setWorked] = useState('');
  const [missing, setMissing] = useState('');
  const [email, setEmail] = useState('');
  const [submitting, setSubmitting] = useState(false);

  async function submit() {
    if (!whatTried.trim() && !worked.trim() && !missing.trim()) {
      toast.error('Add at least one line — what did you try?');
      return;
    }
    setSubmitting(true);
    try {
      await api('/api/feedback', {
        method: 'POST',
        body: JSON.stringify({
          what_tried: whatTried,
          worked,
          missing,
          email,
          page_url: typeof window !== 'undefined' ? window.location.href : '',
        }),
      });
      toast.success('Thanks — we read every one.');
      setOpen(false);
      setWhatTried('');
      setWorked('');
      setMissing('');
      setEmail('');
    } catch (e) {
      toast.error(e instanceof ApiError ? e.message : 'Failed to submit');
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Dialog.Root open={open} onOpenChange={setOpen}>
      <Dialog.Trigger>
        <Button
          size="1"
          variant="soft"
          color="gray"
          className="font-medium"
          aria-label="Send feedback"
        >
          <ChatBubbleIcon />
          {!compact && label}
        </Button>
      </Dialog.Trigger>
      <Dialog.Content maxWidth="520px">
        <Dialog.Title>Send feedback</Dialog.Title>
        <Dialog.Description size="2" color="gray">
          Every submission lands in our inbox. Short is fine — even one field is useful.
        </Dialog.Description>
        <Flex direction="column" gap="3" mt="4">
          <Field label="What did you try?">
            <TextArea
              placeholder="Spun up a scenario, hit the Stripe mock from my SDK…"
              rows={3}
              value={whatTried}
              onChange={(e) => setWhatTried(e.target.value)}
            />
          </Field>
          <Field label="What worked?">
            <TextArea
              placeholder="The X-TestPay-Outcome header was exactly what I wanted."
              rows={2}
              value={worked}
              onChange={(e) => setWorked(e.target.value)}
            />
          </Field>
          <Field label="What's missing or confusing?">
            <TextArea
              placeholder="Couldn't figure out how to…"
              rows={3}
              value={missing}
              onChange={(e) => setMissing(e.target.value)}
            />
          </Field>
          <Field label="Your email (optional — we'll follow up if you want)">
            <TextField.Root
              type="email"
              placeholder="you@example.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
          </Field>
          <Flex gap="2" justify="end" mt="2">
            <Dialog.Close>
              <Button variant="soft" color="gray">
                Cancel
              </Button>
            </Dialog.Close>
            <Button onClick={submit} loading={submitting}>
              Send
            </Button>
          </Flex>
        </Flex>
      </Dialog.Content>
    </Dialog.Root>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <Text as="label" size="2" weight="medium" className="block mb-1">
        {label}
      </Text>
      {children}
    </div>
  );
}
