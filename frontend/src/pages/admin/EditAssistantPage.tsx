import { useNavigate, useParams } from "react-router-dom";
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
import { ApiError, type Assistant } from "@/lib/api";
import { useAssistant } from "@/lib/api/queries/useAssistant";
import { useCategories } from "@/lib/api/queries/useCategories";
import { useUpdateAssistant } from "@/lib/api/queries/useUpdateAssistant";

function toFormValues(a: Assistant): AssistantFormValues {
  return {
    name: a.name,
    description: a.description,
    categoryId: a.categoryId,
    model: a.model,
    systemPrompt: a.systemPrompt ?? "",
    exampleUserPrompt: a.exampleUserPrompt ?? "",
    isActive: a.isActive,
  };
}

export function EditAssistantPage() {
  const { id = "" } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const assistantQuery = useAssistant(id);
  const categoriesQuery = useCategories();
  const updateMutation = useUpdateAssistant(id);

  const handleSubmit = async (values: AssistantFormOut) => {
    try {
      await updateMutation.mutateAsync(values);
      toast.success("Изменения сохранены");
      navigate("/assistants");
    } catch (err) {
      const message =
        err instanceof ApiError ? err.message : "Не удалось обновить ассистента";
      toast.error(message);
    }
  };

  return (
    <section className="flex flex-col gap-8">
      <QueryStateBoundary query={assistantQuery}>
        {(assistant) => (
          <QueryStateBoundary query={categoriesQuery}>
            {(cats) => (
              <>
                <Breadcrumb>
                  <BreadcrumbList>
                    <BreadcrumbItem>
                      <BreadcrumbLink href="/assistants">Каталог</BreadcrumbLink>
                    </BreadcrumbItem>
                    <BreadcrumbSeparator />
                    <BreadcrumbItem>
                      <BreadcrumbLink href={`/assistants/${assistant.id}`}>
                        {assistant.name}
                      </BreadcrumbLink>
                    </BreadcrumbItem>
                    <BreadcrumbSeparator />
                    <BreadcrumbItem>
                      <BreadcrumbPage>Редактирование</BreadcrumbPage>
                    </BreadcrumbItem>
                  </BreadcrumbList>
                </Breadcrumb>

                <header>
                  <p className="text-sm font-semibold text-primary">
                    Администрирование
                  </p>
                  <h1 className="mt-1 text-4xl font-extrabold tracking-tight">
                    Редактирование ассистента
                  </h1>
                </header>

                <AssistantForm
                  mode="edit"
                  defaultValues={toFormValues(assistant)}
                  categories={cats.categories}
                  submitLabel="Сохранить"
                  onSubmit={handleSubmit}
                  onCancel={() => navigate("/assistants")}
                />
              </>
            )}
          </QueryStateBoundary>
        )}
      </QueryStateBoundary>
    </section>
  );
}
