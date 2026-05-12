import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import {
  AssistantForm,
  type AssistantFormOut,
  type AssistantFormValues,
} from "@/components/AssistantForm";
import { QueryStateBoundary } from "@/components/QueryStateBoundary";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { ApiError } from "@/lib/api";
import { useCategories } from "@/lib/api/queries/useCategories";
import { useCreateAssistant } from "@/lib/api/queries/useCreateAssistant";

const defaults: AssistantFormValues = {
  name: "",
  description: "",
  categoryId: "",
  model: "mock-smart",
  systemPrompt: "",
  exampleUserPrompt: "",
  isActive: true,
};

export function NewAssistantPage() {
  const navigate = useNavigate();
  const categoriesQuery = useCategories();
  const createMutation = useCreateAssistant();

  const handleSubmit = async (values: AssistantFormOut) => {
    try {
      await createMutation.mutateAsync(values);
      toast.success("Ассистент создан");
      navigate("/assistants");
    } catch (err) {
      const message =
        err instanceof ApiError ? err.message : "Не удалось создать ассистента";
      toast.error(message);
    }
  };

  return (
    <section className="flex flex-col gap-8">
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbLink href="/assistants">Каталог</BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>Новый ассистент</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>

      <header>
        <p className="text-sm font-semibold text-primary">Администрирование</p>
        <h1 className="mt-1 text-4xl font-extrabold tracking-tight">
          Новый ассистент
        </h1>
      </header>

      <QueryStateBoundary query={categoriesQuery}>
        {(data) => (
          <AssistantForm
            mode="create"
            defaultValues={defaults}
            categories={data.categories}
            submitLabel="Создать"
            onSubmit={handleSubmit}
            onCancel={() => navigate("/assistants")}
          />
        )}
      </QueryStateBoundary>
    </section>
  );
}
