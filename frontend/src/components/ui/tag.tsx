import { mergeProps } from "@base-ui/react/merge-props"
import { useRender } from "@base-ui/react/use-render"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

/**
 * Tag — pill-форма (полностью круглые края), используется для меток
 * категорий и статусов вроде «Новое», «Промо», статусов запусков и т.п.
 */
const tagVariants = cva(
  "group/tag inline-flex h-5 w-fit shrink-0 items-center justify-center gap-1 overflow-hidden rounded-4xl border border-transparent px-2 py-0.5 text-xs font-semibold whitespace-nowrap transition-all focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 has-data-[icon=inline-end]:pr-1.5 has-data-[icon=inline-start]:pl-1.5 aria-invalid:border-destructive aria-invalid:ring-destructive/20 [&>svg]:pointer-events-none [&>svg]:size-3!",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground [a]:hover:bg-primary/80",
        secondary:
          "bg-secondary text-secondary-foreground [a]:hover:bg-secondary/80",
        // Красный pill — «Новое» в навигации Авито
        destructive:
          "bg-destructive text-destructive-foreground focus-visible:ring-destructive/30 [a]:hover:bg-destructive/90",
        // Чёрная pill — активный таб («Все 15»)
        black: "bg-avito-black text-white [a]:hover:bg-avito-black/90",
        // Фиолетовая
        purple:
          "bg-avito-purple text-avito-purple-foreground [a]:hover:bg-avito-purple/90",
        // Зелёная — статус success
        success: "bg-avito-success/15 text-avito-success",
        outline:
          "border-border text-foreground [a]:hover:bg-muted [a]:hover:text-muted-foreground",
        ghost: "hover:bg-muted hover:text-muted-foreground",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
)

function Tag({
  className,
  variant = "default",
  render,
  ...props
}: useRender.ComponentProps<"span"> & VariantProps<typeof tagVariants>) {
  return useRender({
    defaultTagName: "span",
    props: mergeProps<"span">(
      {
        className: cn(tagVariants({ variant }), className),
      },
      props
    ),
    render,
    state: {
      slot: "tag",
      variant,
    },
  })
}

export { Tag, tagVariants }
