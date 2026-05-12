import { Tabs as TabsPrimitive } from "@base-ui/react/tabs"
import { cva, type VariantProps } from "class-variance-authority"

import { cn } from "@/lib/utils"

function Tabs({
  className,
  orientation = "horizontal",
  ...props
}: TabsPrimitive.Root.Props) {
  return (
    <TabsPrimitive.Root
      data-slot="tabs"
      data-orientation={orientation}
      className={cn(
        "group/tabs flex gap-4 data-horizontal:flex-col",
        className
      )}
      {...props}
    />
  )
}

const tabsListVariants = cva(
  "group/tabs-list inline-flex w-fit items-center text-muted-foreground group-data-vertical/tabs:flex-col data-[variant=line]:gap-6 data-[variant=line]:border-b data-[variant=line]:border-border",
  {
    variants: {
      variant: {
        // По умолчанию — line (стиль Авито: жирный текст с подчёркиванием)
        line: "gap-6 border-b border-border bg-transparent",
        // Прежний bg-muted вариант оставим как fallback для pill-табов
        pill: "h-9 gap-1 rounded-xl bg-muted p-1",
      },
    },
    defaultVariants: {
      variant: "line",
    },
  }
)

function TabsList({
  className,
  variant = "line",
  ...props
}: TabsPrimitive.List.Props & VariantProps<typeof tabsListVariants>) {
  return (
    <TabsPrimitive.List
      data-slot="tabs-list"
      data-variant={variant}
      className={cn(tabsListVariants({ variant }), className)}
      {...props}
    />
  )
}

function TabsTrigger({
  className,
  children,
  count,
  ...props
}: TabsPrimitive.Tab.Props & { count?: number | string }) {
  return (
    <TabsPrimitive.Tab
      data-slot="tabs-trigger"
      className={cn(
        "relative inline-flex items-center gap-1.5 whitespace-nowrap text-muted-foreground transition-colors outline-none hover:text-foreground focus-visible:text-foreground disabled:pointer-events-none disabled:opacity-50",
        // line variant (default) — Avito style. Прозрачный border у всех табов,
        // чтобы baseline не сдвигался при активации.
        "group-data-[variant=line]/tabs-list:-mb-px group-data-[variant=line]/tabs-list:pb-3 group-data-[variant=line]/tabs-list:border-b-2 group-data-[variant=line]/tabs-list:border-transparent group-data-[variant=line]/tabs-list:text-xl group-data-[variant=line]/tabs-list:font-bold group-data-[variant=line]/tabs-list:tracking-tight group-data-[variant=line]/tabs-list:data-active:text-foreground group-data-[variant=line]/tabs-list:data-active:border-foreground",
        // pill variant
        "group-data-[variant=pill]/tabs-list:rounded-lg group-data-[variant=pill]/tabs-list:px-3 group-data-[variant=pill]/tabs-list:py-1 group-data-[variant=pill]/tabs-list:text-sm group-data-[variant=pill]/tabs-list:font-semibold group-data-[variant=pill]/tabs-list:data-active:bg-background group-data-[variant=pill]/tabs-list:data-active:text-foreground group-data-[variant=pill]/tabs-list:data-active:shadow-sm",
        "[&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
        className
      )}
      {...props}
    >
      {children}
      {count !== undefined && (
        <sup className="ml-1 text-xs font-bold text-muted-foreground">
          {count}
        </sup>
      )}
    </TabsPrimitive.Tab>
  )
}

function TabsContent({ className, ...props }: TabsPrimitive.Panel.Props) {
  return (
    <TabsPrimitive.Panel
      data-slot="tabs-content"
      className={cn("flex-1 outline-none", className)}
      {...props}
    />
  )
}

export { Tabs, TabsList, TabsTrigger, TabsContent, tabsListVariants }
