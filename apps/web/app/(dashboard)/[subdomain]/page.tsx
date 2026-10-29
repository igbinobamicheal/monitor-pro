import { redirect } from "next/navigation";

export default async function DashboardRootPage({
  params,
}: {
  params: Promise<{ subdomain: string }>;
}) {
  const { subdomain } = await params;
  redirect(`/${subdomain}/overview`);
}
