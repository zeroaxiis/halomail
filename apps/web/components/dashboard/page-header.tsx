export function PageHeader({
  title,
  description,
  action,
}: { title: string; description?: string; action?: React.ReactNode }) {
  return (
    <div className="workspace-page-heading">
      <div>
        <h1>{title}</h1>
        {description && <p>{description}</p>}
      </div>
      {action && <div className="workspace-page-action">{action}</div>}
    </div>
  );
}
