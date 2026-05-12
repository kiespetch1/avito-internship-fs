import { mergeProps } from "@base-ui/react/merge-props"
import { useRender } from "@base-ui/react/use-render"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

/**
 * Badge — info-метка с лёгким скруглением (~6px), для характеристик и инфо
 * вроде «Надёжный партнёр», «3 продажи с Авито Доставкой», «Эковклад».
 * Для pill-скругления (категории, «Новое») используй <Tag>.
 */
const badgeVariants = cva(
  "group/badge inline-flex h-[26px] w-fit shrink-0 items-center justify-center gap-1.5 overflow-hidden rounded-[6px] border border-transparent px-2 text-[0.8125rem] font-normal whitespace-nowrap transition-all focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 aria-invalid:border-destructive aria-invalid:ring-destructive/20 [&>svg]:pointer-events-none [&>svg]:size-3.5",
  {
    variants: {
      variant: {
        // Мягко-голубой info — «Надёжный партнёр», «3 продажи», «Эковклад»
        info: "bg-avito-info-bg text-avito-info-fg [a]:hover:bg-avito-info-bg/80",
        // Серый — нейтральные характеристики
        default:
          "bg-secondary text-secondary-foreground [a]:hover:bg-secondary/80",
        // Чёрный — премиум-метки
        black: "bg-avito-black text-white",
        // Зелёный (приглушённый) — успех
        success: "bg-avito-success/15 text-avito-success",
        // Красный (приглушённый) — алерты/предупреждения
        destructive:
          "bg-destructive/10 text-destructive [a]:hover:bg-destructive/15",
        outline: "border-border text-foreground [a]:hover:bg-muted",
      },
    },
    defaultVariants: {
      variant: "info",
    },
  }
)

function Badge({
  className,
  variant = "info",
  render,
  ...props
}: useRender.ComponentProps<"span"> & VariantProps<typeof badgeVariants>) {
  return useRender({
    defaultTagName: "span",
    props: mergeProps<"span">(
      {
        className: cn(badgeVariants({ variant }), className),
      },
      props
    ),
    render,
    state: {
      slot: "badge",
      variant,
    },
  })
}

export { Badge, badgeVariants }
