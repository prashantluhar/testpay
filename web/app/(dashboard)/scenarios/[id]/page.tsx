import { EditScenarioClient } from './client';

export function generateStaticParams() {
  // Static export fallback — real id is read client-side via useParams.
  return [{ id: '_' }];
}

export const dynamicParams = true;

export default function EditScenarioPage() {
  return <EditScenarioClient />;
}
