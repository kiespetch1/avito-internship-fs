import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";
import type { Pagination as PaginationDTO } from "@/lib/api";

type Props = {
  pagination: PaginationDTO;
  onChange: (page: number) => void;
};

function buildPages(current: number, total: number): Array<number | "ellipsis"> {
  if (total <= 7) {
    return Array.from({ length: total }, (_, i) => i + 1);
  }

  const pages: Array<number | "ellipsis"> = [1];
  const start = Math.max(2, current - 1);
  const end = Math.min(total - 1, current + 1);

  if (start > 2) pages.push("ellipsis");
  for (let p = start; p <= end; p++) pages.push(p);
  if (end < total - 1) pages.push("ellipsis");

  pages.push(total);
  return pages;
}

export function PaginationControl({ pagination, onChange }: Props) {
  const { page, pageSize, total } = pagination;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  if (totalPages <= 1) return null;

  const canPrev = page > 1;
  const canNext = page < totalPages;
  const pages = buildPages(page, totalPages);

  const handleClick = (e: React.MouseEvent, target: number) => {
    e.preventDefault();
    if (target >= 1 && target <= totalPages && target !== page) {
      onChange(target);
    }
  };

  return (
    <Pagination>
      <PaginationContent>
        <PaginationItem>
          <PaginationPrevious
            href="#"
            aria-disabled={!canPrev}
            className={!canPrev ? "pointer-events-none opacity-40" : ""}
            onClick={(e) => handleClick(e, page - 1)}
          />
        </PaginationItem>

        {pages.map((p, i) =>
          p === "ellipsis" ? (
            <PaginationItem key={`e-${i}`}>
              <PaginationEllipsis />
            </PaginationItem>
          ) : (
            <PaginationItem key={p}>
              <PaginationLink
                href="#"
                isActive={p === page}
                onClick={(e) => handleClick(e, p)}
              >
                {p}
              </PaginationLink>
            </PaginationItem>
          ),
        )}

        <PaginationItem>
          <PaginationNext
            href="#"
            aria-disabled={!canNext}
            className={!canNext ? "pointer-events-none opacity-40" : ""}
            onClick={(e) => handleClick(e, page + 1)}
          />
        </PaginationItem>
      </PaginationContent>
    </Pagination>
  );
}
