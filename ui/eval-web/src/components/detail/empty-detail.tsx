export function EmptyDetail(props: { title: string; body: string }) {
  return (
    <div className="empty-detail">
      <p className="eyebrow">Detail</p>
      <h2>{props.title}</h2>
      <p className="muted">{props.body}</p>
    </div>
  );
}
