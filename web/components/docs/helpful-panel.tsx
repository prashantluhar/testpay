import Link from 'next/link';
import { Card, Flex, Text } from '@radix-ui/themes';
import {
  GitHubLogoIcon,
  RocketIcon,
  LightningBoltIcon,
  ChatBubbleIcon,
} from '@radix-ui/react-icons';

// Sticky "helpful" panel that fills the right column under the scroll-spy
// TOC. Static, per-visit — not content-aware. Mirrors the Docusaurus /
// Mintlify pattern where the right rail carries conversion + contribution
// CTAs alongside the auto-built TOC.
export function HelpfulPanel() {
  return (
    <Flex direction="column" gap="3" className="text-sm">
      <Card>
        <Flex direction="column" gap="2" p="2">
          <Flex align="center" gap="2">
            <RocketIcon className="text-[var(--accent-11)]" />
            <Text size="2" weight="medium">
              Try TestPay
            </Text>
          </Flex>
          <Text size="1" color="gray" as="p">
            Spin up a workspace in under a minute. Free, open source, hosted demo on Render.
          </Text>
          <Link
            href="/signup"
            className="inline-flex items-center justify-center rounded-md bg-[var(--accent-9)] px-3 py-1.5 text-[var(--accent-contrast)] hover:bg-[var(--accent-10)] transition-colors text-xs font-medium"
          >
            Get started
          </Link>
        </Flex>
      </Card>

      <Card>
        <Flex direction="column" gap="2" p="2">
          <Flex align="center" gap="2">
            <LightningBoltIcon className="text-[var(--accent-11)]" />
            <Text size="2" weight="medium">
              Quick links
            </Text>
          </Flex>
          <Flex direction="column" gap="1" className="text-xs">
            <QuickLink href="/docs">Getting started</QuickLink>
            <QuickLink href="/docs/failure-modes">28 failure modes</QuickLink>
            <QuickLink href="/docs/scenarios">Scenarios guide</QuickLink>
            <QuickLink href="/docs/api">API reference</QuickLink>
            <QuickLink href="/docs/webhooks">Webhook spec</QuickLink>
          </Flex>
        </Flex>
      </Card>

      <Card>
        <Flex direction="column" gap="2" p="2">
          <Flex align="center" gap="2">
            <GitHubLogoIcon className="text-[var(--gray-11)]" />
            <Text size="2" weight="medium">
              Open source
            </Text>
          </Flex>
          <Text size="1" color="gray" as="p">
            MIT-licensed. Issues and PRs welcome.
          </Text>
          <Link
            href="https://github.com/prashantluhar/testpay"
            target="_blank"
            rel="noreferrer"
            className="text-xs text-[var(--accent-11)] hover:text-[var(--accent-12)] inline-flex items-center gap-1"
          >
            github.com/prashantluhar/testpay ↗
          </Link>
        </Flex>
      </Card>

      <Card>
        <Flex direction="column" gap="2" p="2">
          <Flex align="center" gap="2">
            <ChatBubbleIcon className="text-[var(--gray-11)]" />
            <Text size="2" weight="medium">
              Something missing?
            </Text>
          </Flex>
          <Text size="1" color="gray" as="p">
            Open a GitHub issue with the gateway or behavior you&apos;d like to see — fastest way to get it added.
          </Text>
          <Link
            href="https://github.com/prashantluhar/testpay/issues/new"
            target="_blank"
            rel="noreferrer"
            className="text-xs text-[var(--accent-11)] hover:text-[var(--accent-12)]"
          >
            File an issue ↗
          </Link>
        </Flex>
      </Card>
    </Flex>
  );
}

function QuickLink({ href, children }: { href: string; children: React.ReactNode }) {
  return (
    <Link
      href={href}
      className="text-[var(--gray-11)] hover:text-[var(--accent-11)] transition-colors"
    >
      {children}
    </Link>
  );
}
