'use client';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Button } from '@/components/ui/button';
import { useMe } from '@/lib/hooks';
import { ApiKeyReveal } from '@/components/common/api-key-reveal';
import { useTheme } from '@/components/common/theme-provider';
import { CopyButton } from '@/components/common/copy-button';
import { MODE } from '@/lib/types';

export default function SettingsPage() {
  const { data } = useMe();
  const { theme, setTheme } = useTheme();

  if (!data) return null;
  const { workspace, user } = data;

  const baseUrl =
    MODE === 'local' ? 'http://localhost:7700' : `https://api.testpay.dev/ws_${workspace.Slug}`;

  return (
    <div className="max-w-2xl space-y-6">
      <h1 className="text-2xl font-semibold">Settings</h1>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Workspace</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div>
            <Label>Slug</Label>
            <code className="block font-mono text-sm bg-muted px-3 py-2 rounded-md mt-1">
              {workspace.Slug}
            </code>
          </div>
          <div>
            <Label>API key</Label>
            <div className="mt-1">
              <ApiKeyReveal value={workspace.APIKey} />
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Endpoints</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2">
          {['stripe', 'razorpay', 'agnostic'].map((g) => {
            const url = g === 'agnostic' ? `${baseUrl}/v1` : `${baseUrl}/${g}`;
            return (
              <div key={g} className="flex items-center gap-3">
                <span className="w-20 text-sm uppercase text-muted-foreground">{g}</span>
                <code className="flex-1 font-mono text-sm bg-muted px-3 py-2 rounded-md">
                  {url}
                </code>
                <CopyButton value={url} label="" />
              </div>
            );
          })}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Appearance</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-2">
            {(['light', 'dark', 'system'] as const).map((t) => (
              <Button
                key={t}
                variant={theme === t ? 'default' : 'outline'}
                size="sm"
                onClick={() => setTheme(t)}
              >
                {t}
              </Button>
            ))}
          </div>
        </CardContent>
      </Card>

      {MODE === 'hosted' && user && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Account</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-sm text-muted-foreground">{user.Email}</div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
