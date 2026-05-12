import { useId, type ReactNode } from "react";
import { Label } from "@/components/ui/label";

type RenderProps = {
  id: string;
  "aria-invalid": boolean | undefined;
  "aria-describedby": string | undefined;
};

type Props = {
  label?: ReactNode;
  description?: ReactNode;
  error?: string;
  children: (props: RenderProps) => ReactNode;
  className?: string;
};

export function FormField({ label, description, error, children, className }: Props) {
  const id = useId();
  const errorId = `${id}-error`;
  const descId = `${id}-desc`;

  const describedBy =
    error !== undefined ? errorId : description !== undefined ? descId : undefined;

  return (
    <div className={`space-y-2 ${className ?? ""}`}>
      {label !== undefined && (
        <Label htmlFor={id} className="text-[0.9375rem] font-bold">
          {label}
        </Label>
      )}
      {children({
        id,
        "aria-invalid": error !== undefined ? true : undefined,
        "aria-describedby": describedBy,
      })}
      {description !== undefined && error === undefined && (
        <p id={descId} className="text-xs text-muted-foreground">
          {description}
        </p>
      )}
      {error !== undefined && (
        <p id={errorId} className="text-xs text-destructive">
          {error}
        </p>
      )}
    </div>
  );
}
