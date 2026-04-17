import Link from 'next/link';
import { Badge, Card, Flex, Heading, Separator, Table, Text } from '@radix-ui/themes';
import {
  LightningBoltIcon,
  CheckIcon,
  RocketIcon,
  PersonIcon,
  GearIcon,
  BarChartIcon,
  MagicWandIcon,
} from '@radix-ui/react-icons';

// Public "about" page — mirrors the README's positioning section so visitors
// without an account can still understand what the project is, who it's for,
// and how it helps.
export default function AboutPage() {
  return (
    <div className="space-y-8 animate-in fade-in duration-300">
      <section>
        <Flex align="center" gap="3" mb="2">
          <span className="flex h-10 w-10 items-center justify-center rounded-lg bg-[var(--accent-a5)] text-[var(--accent-11)]">
            <LightningBoltIcon width="22" height="22" />
          </span>
          <Heading size="8" weight="bold">
            TestPay
          </Heading>
        </Flex>
        <Text size="4" color="gray" className="block mb-4">
          Postman for Payments.
        </Text>
        <Text size="3" as="p" className="leading-relaxed">
          A mock payment gateway and failure-simulation tool that lets developers test every
          real-world payment edge case — locally and in CI — without touching production systems.
          Drop-in compatible with Stripe, Razorpay, Adyen, Mastercard, and eight more. Single Go
          binary, embedded dashboard, zero external dependencies.
        </Text>
      </section>

      <Separator size="4" />

      <section>
        <Heading size="6" mb="3">
          Why it exists
        </Heading>
        <Flex direction="column" gap="2" className="text-[var(--gray-12)]">
          <Text size="3" as="p">
            Sandbox environments from real PSPs never replicate the failure modes that cause
            incidents in production — bank timeouts, duplicate webhooks, async state
            transitions, 3DS redirect abandonment, rate-limiting cascades. Most integrations ship
            with the unhappy paths untested because there&apos;s no way to trigger them on demand.
          </Text>
          <Text size="3" as="p">
            TestPay gives you a gateway that behaves exactly like the real thing, including{' '}
            <strong>every way it can fail</strong>. 28 failure modes across 13 gateways, each
            replayable from a URL or a one-line header.
          </Text>
        </Flex>
      </section>

      <section>
        <Heading size="6" mb="3">
          Who it&apos;s for
        </Heading>
        <Flex direction="column" gap="3">
          <Persona
            icon={<GearIcon />}
            title="Payment integration engineers"
            body="Wiring Stripe / Razorpay / Adyen / any PSP for the first time, or debugging a
            production integration. Reproduce the specific failure mode, in seconds, instead of
            rummaging through the gateway's sandbox config."
          />
          <Persona
            icon={<CheckIcon />}
            title="QA and test automation"
            body="Every failure mode becomes a replayable fixture your CI can assert against, so
            the 'this worked yesterday' class of bug stops shipping."
          />
          <Persona
            icon={<RocketIcon />}
            title="DevEx / platform teams"
            body="Internal staging environments where upstream sandbox flakiness (test-mode
            outages, regional DNS blips, rate limits) breaks builds and burns on-call hours."
          />
          <Persona
            icon={<MagicWandIcon />}
            title="Founders and PMs"
            body="Validate the feasibility of a payments integration before committing
            engineering weeks. Spin up TestPay, hit it from a prototype, sanity-check edge cases."
          />
          <Persona
            icon={<PersonIcon />}
            title="Security and compliance reviewers"
            body="See how the integration behaves on bank-level declines and CVV mismatches
            without logging in to production."
          />
        </Flex>
      </section>

      <section>
        <Heading size="6" mb="3">
          How it saves time
        </Heading>
        <Card>
          <Table.Root size="2" variant="ghost">
            <Table.Header>
              <Table.Row>
                <Table.ColumnHeaderCell>Pain today</Table.ColumnHeaderCell>
                <Table.ColumnHeaderCell>With TestPay</Table.ColumnHeaderCell>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              <SaveRow
                pain={'"How do I force a bank timeout?" — wait for the sandbox to happen to do it, or open a support ticket'}
                win={'Run scenario run bank-timeout — the exact failure fires every time'}
              />
              <SaveRow
                pain={'CI depends on the PSP test mode. Build broke because their sandbox is rate-limited.'}
                win={'Local binary. No external dependency, no flaky builds.'}
              />
              <SaveRow
                pain={'Shipped a bug that only triggers on duplicate webhook — can\'t reproduce locally.'}
                win={'webhook_duplicate fires on demand. One click to replay.'}
              />
              <SaveRow
                pain={'"3DS cancel" bug report. Never been able to trigger one in dev.'}
                win={'redirect_abandoned scenario. 50 ms.'}
              />
              <SaveRow
                pain={'New engineer needs to test the integration. Doesn\'t have test API keys.'}
                win={'One binary. No accounts. Same env for everyone.'}
              />
            </Table.Body>
          </Table.Root>
        </Card>
      </section>

      <section>
        <Heading size="6" mb="3">
          How it helps day-to-day
        </Heading>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <FeatureCard title="28 failure modes" body="bank, PG, webhook, redirect/3DS, charge anomalies, async state transitions — all reachable via a single X-TestPay-Outcome header" />
          <FeatureCard title="Named scenarios" body="Save sequences of failure modes as replayable test fixtures — multi-step sequences advance across SDK calls automatically" />
          <FeatureCard title="Full logging" body="Every request, header, response, and webhook delivery logged to Postgres with 8 KB of per-attempt response body captured" />
          <FeatureCard title="Webhook debugger" body="Inspect delivery attempts, retry history, payloads — know exactly what your endpoint saw on each try" />
          <FeatureCard title="Zero code change" body="Point your SDK at localhost:7700/stripe and it just works — drop-in replacement for Stripe, Razorpay, Adyen, 10 more" />
          <FeatureCard title="Single binary" body="Embedded dashboard + API, Docker image under 20 MB, runs on Render free tier with no cloud bill" />
        </div>
      </section>

      <section>
        <Heading size="6" mb="3">
          Cost efficiency
        </Heading>
        <Flex direction="column" gap="3">
          <CostRow label="$0 self-hosted" body="Single Go binary + Postgres. Drops into any host. No per-seat or per-environment licensing." />
          <CostRow label="$0 hosted demo" body="Render + Neon free tiers. Enough for team demos, hackathons, portfolios." />
          <CostRow label="Zero external API burn" body="Your CI and local dev never hit a real gateway's sandbox — no rate-limit bites, no paid simulation tools like WireMock Cloud." />
          <CostRow label="Replaces two categories" body="Payment sandboxes (handled by PSPs but flaky) and generic HTTP mocks (need to hand-roll each response shape). TestPay ships production-accurate gateway shapes out of the box." />
          <CostRow label="Faster onboarding" body="A new engineer is productive in 5 minutes, no test credentials to provision, no MFA on a shared sandbox account." />
        </Flex>
      </section>

      <Separator size="4" />

      <section>
        <Heading size="6" mb="3">
          How to start
        </Heading>
        <Flex direction="column" gap="3">
          <NextStep
            step="1"
            title="Skim the getting-started page"
            href="/docs"
            desc="30 seconds. Gives you the one-line curl that proves the mock is alive."
          />
          <NextStep
            step="2"
            title="Point your SDK at the mock"
            href="/docs/point-your-sdk"
            desc="Stripe SDK, Razorpay, Adyen — each has a copy-pasteable base-URL override snippet."
          />
          <NextStep
            step="3"
            title="Browse the 28 failure modes"
            href="/docs/failure-modes"
            desc="Pick the one you want to reproduce; grab its wire name; drop it in an X-TestPay-Outcome header."
          />
          <NextStep
            step="4"
            title="Create an account to keep your logs"
            href="/signup"
            desc="Free. Gives you a persistent workspace + dashboard + webhook debugger."
          />
        </Flex>
      </section>

      <section>
        <Heading size="6" mb="3">
          Open source
        </Heading>
        <Text size="3" as="p">
          MIT-licensed, source on{' '}
          <Link
            href="https://github.com/prashantluhar/testpay"
            target="_blank"
            className="underline text-[var(--accent-11)]"
          >
            github.com/prashantluhar/testpay
          </Link>
          . Self-host it with one Go build, or use the hosted demo — both free.{' '}
          <Link href="/docs" className="underline text-[var(--accent-11)]">
            Read the docs →
          </Link>
        </Text>
      </section>
    </div>
  );
}

function Persona({ icon, title, body }: { icon: React.ReactNode; title: string; body: string }) {
  return (
    <Flex gap="3" align="start">
      <span className="mt-1 flex h-7 w-7 items-center justify-center rounded-md bg-[var(--gray-a3)] text-[var(--gray-11)] shrink-0">
        {icon}
      </span>
      <div>
        <Text size="3" weight="medium" className="block">
          {title}
        </Text>
        <Text size="2" color="gray" as="p" className="mt-1">
          {body}
        </Text>
      </div>
    </Flex>
  );
}

function SaveRow({ pain, win }: { pain: string; win: string }) {
  return (
    <Table.Row>
      <Table.Cell className="align-top py-3 text-[var(--gray-11)]">{pain}</Table.Cell>
      <Table.Cell className="align-top py-3 text-[var(--gray-12)]">{win}</Table.Cell>
    </Table.Row>
  );
}

function FeatureCard({ title, body }: { title: string; body: string }) {
  return (
    <Card>
      <Flex direction="column" gap="1" p="1">
        <Flex align="center" gap="2">
          <CheckIcon className="text-[var(--accent-11)]" />
          <Text size="3" weight="medium">
            {title}
          </Text>
        </Flex>
        <Text size="2" color="gray" as="p">
          {body}
        </Text>
      </Flex>
    </Card>
  );
}

function CostRow({ label, body }: { label: string; body: string }) {
  return (
    <Flex gap="3" align="start">
      <Badge color="grass" variant="soft" className="shrink-0">
        {label}
      </Badge>
      <Text size="2" as="p" className="text-[var(--gray-12)]">
        {body}
      </Text>
    </Flex>
  );
}

function NextStep({ step, title, href, desc }: { step: string; title: string; href: string; desc: string }) {
  return (
    <Link
      href={href}
      className="group block rounded-lg border border-[var(--gray-a5)] p-4 hover:border-[var(--accent-a8)] hover:bg-[var(--accent-a2)] transition-colors"
    >
      <Flex gap="3" align="start">
        <span className="flex h-7 w-7 items-center justify-center rounded-full bg-[var(--accent-a5)] text-[var(--accent-11)] font-mono text-sm shrink-0">
          {step}
        </span>
        <div>
          <Flex align="center" gap="2">
            <Text size="3" weight="medium">
              {title}
            </Text>
            <BarChartIcon className="opacity-0 group-hover:opacity-70 transition-opacity text-[var(--accent-11)]" />
          </Flex>
          <Text size="2" color="gray" as="p" className="mt-1">
            {desc}
          </Text>
        </div>
      </Flex>
    </Link>
  );
}
