export const dynamic = "force-dynamic";

export async function GET() {
  const apiBase = process.env.API_BASE || "";
  const aiEnabledValue = process.env.ENABLE_AI || "";
  const aiEnabled = String(aiEnabledValue).toLowerCase() === "true";
  return Response.json({ apiBase, aiEnabled });
}
