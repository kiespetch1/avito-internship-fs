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
  // Если page из URL вышел за пределы totalPages (фильтры/удаления сузили выдачу),
  // отображаем как будто пользователь на последней странице. Сам page в URL не трогаем —
  // он подровняется когда пользователь кликнет назад/последнюю
  const displayPage = Math.min(Math.max(1, page), totalPages);
  if (totalPages <= 1) return null;

  const canPrev = displayPage > 1;
  const canNext = displayPage < totalPages;
  const pages = buildPages(displayPage, totalPages);

  const handleClick = (e: React.MouseEvent, target: number) => {
    e.preventDefault();
    if (target >= 1 && target <= totalPages && target !== displayPage) {
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
            onClick={(e) => handleClick(e, displayPage - 1)}
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
                isActive={p === displayPage}
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
            onClick={(e) => handleClick(e, displayPage + 1)}
          />
        </PaginationItem>
      </PaginationContent>
    </Pagination>
  );
}
