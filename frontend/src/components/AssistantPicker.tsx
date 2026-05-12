import { Combobox } from "@base-ui/react/combobox";
import { CheckIcon, ChevronDownIcon, XIcon } from "lucide-react";
import { useMemo, useState } from "react";
import { useAssistant } from "@/lib/api/queries/useAssistant";
import { useAssistantsList } from "@/lib/api/queries/useAssistantsList";
import { useDebouncedValue } from "@/lib/hooks/useDebouncedValue";
import { cn } from "@/lib/utils";

type AssistantOption = { value: string; label: string };

type Props = {
  value: string | null;
  onChange: (value: string | null) => void;
  placeholder?: string;
  className?: string;
};

export function AssistantPicker({
  value,
  onChange,
  placeholder = "Все ассистенты",
  className,
}: Props) {
  const [query, setQuery] = useState("");
  const debouncedQuery = useDebouncedValue(query, 300);

  const listQuery = useAssistantsList({
    q: debouncedQuery.trim() === "" ? undefined : debouncedQuery.trim(),
    includeInactive: true,
    page: 1,
    pageSize: 20,
  });

  const selectedAssistantQuery = useAssistant(value ?? "");

  const items = useMemo<AssistantOption[]>(() => {
    const list = listQuery.data?.assistants ?? [];
    return list.map((a) => ({ value: a.id, label: a.name }));
  }, [listQuery.data]);

  const selectedOption = useMemo<AssistantOption | null>(() => {
    if (value === null) return null;
    const inList = items.find((i) => i.value === value);
    if (inList) return inList;
    const fetched = selectedAssistantQuery.data;
    return { value, label: fetched?.name ?? "…" };
  }, [value, items, selectedAssistantQuery.data]);

  return (
    <Combobox.Root<AssistantOption, false>
      items={items}
      filter={null}
      autoComplete="none"
      value={selectedOption}
      onValueChange={(next) => onChange(next === null ? null : next.value)}
      onInputValueChange={(input) => setQuery(input)}
    >
      <Combobox.InputGroup
        className={cn(
          "relative flex h-12 w-full items-center gap-2 rounded-lg border border-transparent bg-field px-3 text-[0.9375rem] transition-colors focus-within:border-ring focus-within:ring-3 focus-within:ring-ring/30",
          className,
        )}
      >
        <Combobox.Input
          placeholder={placeholder}
          className="flex-1 bg-transparent outline-none placeholder:text-muted-foreground"
        />
        {value !== null && (
          <Combobox.Clear
            className="rounded p-1 text-muted-foreground hover:text-foreground"
            aria-label="Очистить"
          >
            <XIcon className="size-4" />
          </Combobox.Clear>
        )}
        <Combobox.Trigger
          className="rounded p-1 text-muted-foreground hover:text-foreground"
          aria-label="Раскрыть список"
        >
          <ChevronDownIcon className="size-4" />
        </Combobox.Trigger>
      </Combobox.InputGroup>

      <Combobox.Portal>
        <Combobox.Positioner sideOffset={8} className="isolate z-50">
          <Combobox.Popup className="max-h-72 w-(--anchor-width) min-w-56 overflow-y-auto rounded-xl bg-popover p-1 text-popover-foreground shadow-lg ring-1 ring-foreground/5">
            <Combobox.Empty className="px-3 py-2 text-sm text-muted-foreground">
              {listQuery.isFetching ? "Загрузка..." : "Ничего не найдено"}
            </Combobox.Empty>
            <Combobox.List>
              {(item: AssistantOption) => (
                <Combobox.Item
                  key={item.value}
                  value={item}
                  className="relative flex cursor-default items-center gap-2 rounded-lg px-3 py-2 text-[0.9375rem] outline-none data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground"
                >
                  <span className="flex-1">{item.label}</span>
                  <Combobox.ItemIndicator>
                    <CheckIcon className="size-4" />
                  </Combobox.ItemIndicator>
                </Combobox.Item>
              )}
            </Combobox.List>
          </Combobox.Popup>
        </Combobox.Positioner>
      </Combobox.Portal>
    </Combobox.Root>
  );
}
