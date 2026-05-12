"use client"

import { Checkbox as CheckboxPrimitive } from "@base-ui/react/checkbox"
import { CheckboxGroup as CheckboxGroupPrimitive } from "@base-ui/react/checkbox-group"
import { CheckIcon, MinusIcon } from "lucide-react"

import { cn } from "@/lib/utils"

function CheckboxGroup({
  className,
  ...props
}: CheckboxGroupPrimitive.Props) {
  return (
    <CheckboxGroupPrimitive
      data-slot="checkbox-group"
      className={cn("grid w-full gap-3", className)}
      {...props}
    />
  )
}

function Checkbox({ className, ...props }: CheckboxPrimitive.Root.Props) {
  return (
    <CheckboxPrimitive.Root
      data-slot="checkbox"
      className={cn(
        "peer group/checkbox relative inline-flex size-5 shrink-0 cursor-pointer items-center justify-center rounded-[5px] bg-[#d1d3d5] text-white outline-none transition-colors focus-visible:ring-3 focus-visible:ring-ring/30 disabled:cursor-not-allowed disabled:opacity-50 aria-invalid:bg-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 data-checked:bg-avito-black data-indeterminate:bg-avito-black",
        className
      )}
      {...props}
    >
      <CheckboxPrimitive.Indicator
        data-slot="checkbox-indicator"
        className="flex items-center justify-center text-white"
      >
        <CheckIcon
          className="size-3.5 group-data-indeterminate/checkbox:hidden"
          strokeWidth={3}
        />
        <MinusIcon
          className="hidden size-3.5 group-data-indeterminate/checkbox:block"
          strokeWidth={3}
        />
      </CheckboxPrimitive.Indicator>
    </CheckboxPrimitive.Root>
  )
}

export { Checkbox, CheckboxGroup }
